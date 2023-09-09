package bot

import (
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
	HistoryDepth         = 10
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
	SpawnRate     int
	VariableDelay int
	Finished      bool
	LastMessages  History
	Winner        string
}

func (g *Game) Spawns(rng U.RNG) bool {
	if !g.LastMessages.HasSeveralAuthors() || g.LastMessages.Span() > time.Hour {
		// if g.LastMessages.Span() > time.Hour { // for debug
		return false
	}
	return rng.PercentChance(g.SpawnRate)
}

func (g *Game) IncreaseSpawnRate(rng U.RNG, msg *DG.Message) {
	g.LastMessages = g.LastMessages.Update(msg)
	span := g.LastMessages.Span()
	if span > time.Hour {
		g.ResetSpawnRate()
		g.SpawnRate += rng.Intn(2)
	} else if g.LastMessages.Span() > 15*time.Minute {
		return
	} else {
		g.SpawnRate += rng.Intn(3)
	}
}

func (g *Game) ResetSpawnRate() {
	g.SpawnRate = 0
}

type Server struct {
	ID       string
	Prefix   string
	G        Game
	Channels []string
	Admins   []string
	Users    map[string][]string
	Lb       Leaderboard
}

// CanSpawn returns true only if an item can spawn in the given channel
func (s Server) CanSpawn(cid string) bool {
	_, channelHasSpawn := s.G.Monsters[cid]
	// If the game is not on, the channel is not a spawn channel, or the channel already has
	// a spawn, a spawn cannot happen
	if !s.G.On || !U.Contains(s.Channels, cid) || channelHasSpawn {
		return false
	}
	// cooldown check
	delayOk := time.Now().Local().After(s.G.NextSpawn)
	return delayOk
}

// Cooldown sets the server game on cooldown
func (s *Server) Cooldown(rng U.RNG) {
	randomDelay := time.Duration(rng.Intn(s.G.VariableDelay)) * time.Second
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
	Log                 *LR.Logger
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
	rng                 U.RNG
}

type BotAction func(*Bot, CommandParameters)

type Command struct {
	Name           string
	Action         BotAction
	appCmd         *DG.ApplicationCommand
	Options        Options
	Admin          bool
	AlwaysTrigger  bool
	ModifiesServer bool
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
		Name:           "trick",
		Action:         Grab,
		Options:        Options{},
		ModifiesServer: true,
	},
	{
		Name:           "treat",
		Action:         Grab,
		Options:        Options{},
		ModifiesServer: true,
	},
	{
		Name:           "spawn",
		Action:         Spawn,
		appCmd:         &DG.ApplicationCommand{Description: "Forces a random monster to appear (for testing purposes)"},
		Options:        Options{},
		AlwaysTrigger:  true,
		ModifiesServer: true,
	},

	// Score commands
	{
		Name:           "leaderboard",
		Action:         ShowLeaderboard,
		appCmd:         &DG.ApplicationCommand{Description: "Show the server leaderboard"},
		Options:        Options{},
		ModifiesServer: true,
	},
	{
		Name:           "score",
		Action:         ShowScore,
		appCmd:         &DG.ApplicationCommand{Description: "Get your scoreboard"},
		Options:        Options{},
		ModifiesServer: true,
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
		Name:           "setprefix",
		Action:         SetPrefix,
		appCmd:         &DG.ApplicationCommand{Description: "Change the bot prefix"},
		Options:        Options{{"prefix", "new prefix to use", TypeString}},
		Admin:          true,
		ModifiesServer: true,
	},
	{
		Name:   "setcd",
		Action: SetCooldown,
		appCmd: &DG.ApplicationCommand{Description: "Change the spawn cooldown"},
		Options: Options{
			{"minimum", "minimum cooldown duration (in seconds)", TypeInteger},
			{"maximum", "maximum cooldown duration (in seconds)", TypeInteger},
		},
		Admin:          true,
		ModifiesServer: true,
	},
	{
		Name:           "setstay",
		Action:         SetStay,
		appCmd:         &DG.ApplicationCommand{Description: "Change how long a monsters stays idle"},
		Options:        Options{{"duration", "duration in seconds", TypeInteger}},
		Admin:          true,
		ModifiesServer: true,
	},
	{
		Name:           "addchan",
		Action:         AddChannel,
		appCmd:         &DG.ApplicationCommand{Description: "Add a channel where monsters can spawn"},
		Options:        Options{{"channel", "channel to add", TypeChannel}},
		Admin:          true,
		ModifiesServer: true,
	},
	{
		Name:           "rmvchan",
		Action:         RemoveChannel,
		appCmd:         &DG.ApplicationCommand{Description: "Remove a channel where monsters can spawn"},
		Options:        Options{{"channel", "channel to remove", TypeChannel}},
		Admin:          true,
		ModifiesServer: true,
	},
	{
		Name:    "chanlist",
		Action:  Channels,
		appCmd:  &DG.ApplicationCommand{Description: "Display the list of channels where monsters can spawn"},
		Options: Options{},
		Admin:   true,
	},
	{
		Name:           "addadmin",
		Action:         AddAdmin,
		appCmd:         &DG.ApplicationCommand{Description: "Add a bot administrator"},
		Options:        Options{{"user", "user that shall be an administrator", TypeUser}},
		Admin:          true,
		ModifiesServer: true,
	},
	{
		Name:           "rmvadmin",
		Action:         RemoveAdmin,
		appCmd:         &DG.ApplicationCommand{Description: "Remove a bot administrator"},
		Options:        Options{{"user", "user that be removed from the administrators list", TypeUser}},
		Admin:          true,
		ModifiesServer: true,
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
		Name:           "play",
		Action:         Play,
		appCmd:         &DG.ApplicationCommand{Description: "Start or stop the game"},
		Options:        Options{{"state", "game status: \"on\" or \"off\"", TypeString}},
		Admin:          true,
		ModifiesServer: true,
	},
	{
		Name:           "reset",
		Action:         Reset,
		appCmd:         &DG.ApplicationCommand{Description: "Reset the game for all users in the server"},
		Options:        Options{},
		Admin:          true,
		ModifiesServer: true,
	},
	{
		Name:           "give",
		Action:         GiveNew,
		appCmd:         &DG.ApplicationCommand{Description: "Give a random item (for testing purposes)"},
		Options:        Options{},
		Admin:          true,
		ModifiesServer: true,
	},
}

type InteractionHandlers map[string]func(*DG.Session, *DG.InteractionCreate)
