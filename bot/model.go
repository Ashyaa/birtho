package bot

import (
	"math/rand"
	"sync"
	"time"

	"github.com/asdine/storm/v3"
	U "github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
	LR "github.com/sirupsen/logrus"
)

const (
	DefaultPrefix        = "b!"
	DefaultMinDelay      = 120
	DefaultVariableDelay = 781
	DefaultStayTime      = 5 * time.Second
)

var DefaultMemberPermissions int64 = DG.PermissionManageServer

type MonsterSpawn struct {
	ID       string
	Message  string
	Expected string
}

type Game struct {
	On            bool
	Monsters      map[string]MonsterSpawn
	NextSpawn     time.Time
	MinDelay      time.Duration
	StayTime      time.Duration
	VariableDelay int
	Finished      bool
	Winner        string
}

type Server struct {
	ID       string
	Prefix   string
	G        Game
	Channels []string
	Admins   []string
	Users    map[string][]string
}

// CanSpawn returns true only if an item can spawn in the given channel
func (s Server) CanSpawn(cid string) bool {
	delayOk := time.Now().Local().After(s.G.NextSpawn)
	return s.G.On && delayOk && U.Contains(s.Channels, cid)
}

// Cooldown sets the server game on cooldown
func (s *Server) Cooldown() {
	randomDelay := time.Duration(rand.Intn(s.G.VariableDelay)) * time.Second
	s.G.NextSpawn = time.Now().Local().Add(s.G.MinDelay) // Base cooldown of 2mn
	s.G.NextSpawn = s.G.NextSpawn.Add(randomDelay)       // variable cooldown, up to 15mn total
}

func (s Server) IsAdmin(uid string) bool {
	// Everyone is an admin when no admins are registered
	if len(s.Admins) == 0 {
		return true
	}
	return U.Contains(s.Admins, uid)
}

type Bot struct {
	s                   *DG.Session
	db                  *storm.DB
	Log                 LR.Logger
	UserID              string
	Mention             string
	Items               map[string]Item
	Monsters            map[string]Monster
	MonsterIds          []string
	Menus               map[string]Menu
	EqualMonsterChances bool
	InteractionHandlers InteractionHandlers
	Commands            []Command
	mutex               sync.Mutex
}

type BotAction func(*Bot, CommandParameters)

type Command struct {
	Name          string
	Action        BotAction
	appCmd        *DG.ApplicationCommand
	Options       Options
	Admin         bool
	AlwaysTrigger bool
}

func (c Command) ID() string {
	if c.appCmd == nil {
		return ""
	}
	return c.appCmd.ID
}

var commandList = []Command{
	// Game commands
	{
		Name:    "trick",
		Action:  Grab,
		Options: Options{},
	},
	{
		Name:    "treat",
		Action:  Grab,
		Options: Options{},
	},
	{
		Name:          "spawn",
		Action:        Spawn,
		appCmd:        &DG.ApplicationCommand{Description: "Forces a random monster to appear (for testing purposes)"},
		Options:       Options{},
		AlwaysTrigger: true,
	},

	// Score commands
	{
		Name:    "leaderboard",
		Action:  Leaderboard,
		appCmd:  &DG.ApplicationCommand{Description: "Show the server leaderboard"},
		Options: Options{},
	},
	{
		Name:    "score",
		Action:  Score,
		appCmd:  &DG.ApplicationCommand{Description: "Get your scoreboard"},
		Options: Options{},
	},

	// Help command
	{
		Name:    "help",
		Action:  Help,
		appCmd:  &DG.ApplicationCommand{Description: "Information about the bot and how to play the game"},
		Options: Options{},
	},

	// Administration commands
	{
		Name:    "prefix",
		Action:  Prefix,
		appCmd:  &DG.ApplicationCommand{Description: "Display the bot prefix"},
		Options: Options{},
		Admin:   true,
	},
	{
		Name:    "setprefix",
		Action:  SetPrefix,
		appCmd:  &DG.ApplicationCommand{Description: "Change the bot prefix"},
		Options: Options{{"prefix", "new prefix to use", TypeString}},
		Admin:   true,
	},
	{
		Name:   "setcd",
		Action: SetCooldown,
		appCmd: &DG.ApplicationCommand{Description: "Change the spawn cooldown"},
		Options: Options{
			{"minimum", "minimum cooldown duration (in seconds)", TypeInteger},
			{"maximum", "maximum cooldown duration (in seconds)", TypeInteger},
		},
		Admin: true,
	},
	{
		Name:    "setstay",
		Action:  SetStay,
		appCmd:  &DG.ApplicationCommand{Description: "Change how long a monsters stays idle"},
		Options: Options{{"duration", "duration in seconds", TypeInteger}},
		Admin:   true,
	},
	{
		Name:    "addchan",
		Action:  AddChannel,
		appCmd:  &DG.ApplicationCommand{Description: "Add a channel where monsters can spawn"},
		Options: Options{{"channel", "channel to add", TypeChannel}},
		Admin:   true,
	},
	{
		Name:    "rmvchan",
		Action:  RemoveChannel,
		appCmd:  &DG.ApplicationCommand{Description: "Remove a channel where monsters can spawn"},
		Options: Options{{"channel", "channel to remove", TypeChannel}},
		Admin:   true,
	},
	{
		Name:    "chanlist",
		Action:  Channels,
		appCmd:  &DG.ApplicationCommand{Description: "Display the list of channels where monsters can spawn"},
		Options: Options{},
		Admin:   true,
	},
	{
		Name:    "addadmin",
		Action:  AddAdmin,
		appCmd:  &DG.ApplicationCommand{Description: "Add a bot administrator"},
		Options: Options{{"user", "user that shall be an administrator", TypeUser}},
		Admin:   true,
	},
	{
		Name:    "rmvadmin",
		Action:  RemoveAdmin,
		appCmd:  &DG.ApplicationCommand{Description: "Remove a bot administrator"},
		Options: Options{{"user", "user that be removed from the administrators list", TypeUser}},
		Admin:   true,
	},
	{
		Name:    "admins",
		Action:  Admins,
		appCmd:  &DG.ApplicationCommand{Description: "Display the list of bot administrators"},
		Options: Options{},
		Admin:   true,
	},
	{
		Name:    "info",
		Action:  Info,
		appCmd:  &DG.ApplicationCommand{Description: "Information about the bot configuration"},
		Options: Options{},
		Admin:   true,
	},
	{
		Name:    "play",
		Action:  Play,
		appCmd:  &DG.ApplicationCommand{Description: "Start or stop the game"},
		Options: Options{{"state", "game status: \"on\" or \"off\"", TypeString}},
		Admin:   true,
	},
	{
		Name:    "reset",
		Action:  Reset,
		appCmd:  &DG.ApplicationCommand{Description: "Reset the game for all users in the server"},
		Options: Options{},
		Admin:   true,
	},
	{
		Name:    "give",
		Action:  GiveRandom,
		appCmd:  &DG.ApplicationCommand{Description: "Give a random item (for testing purposes)"},
		Options: Options{},
		Admin:   true,
	},
}

type InteractionHandlers map[string]func(*DG.Session, *DG.InteractionCreate)
