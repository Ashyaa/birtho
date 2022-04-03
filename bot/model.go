package bot

import (
	"math/rand"
	"time"

	"github.com/asdine/storm/v3"
	U "github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
	LR "github.com/sirupsen/logrus"
)

const DefaultPrefix = "b!"

type ItemSpawn struct {
	ID      string
	Message string
}

type Game struct {
	On        bool
	Items     map[string]ItemSpawn
	messageID string
	NextSpawn time.Time
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
	s.G.NextSpawn = time.Now().Local().Add(2 * time.Minute)                        // Base cooldown of 2mn
	s.G.NextSpawn = s.G.NextSpawn.Add(time.Duration(rand.Intn(780)) * time.Second) // variable cooldown, up to 15mn total
}

func (s Server) IsAdmin(uid string) bool {
	// Everyone is an admin when no admins are registered
	if len(s.Admins) == 0 {
		return true
	}
	return U.Contains(s.Admins, uid)
}

type Bot struct {
	dg                  *DG.Session
	db                  *storm.DB
	Log                 LR.Logger
	UserID              string
	Mention             string
	Monsters            map[string]Monster
	MonsterIds          []string
	EqualMonsterChances bool
}

type HandlerConstructor func(*Bot, string) func(*DG.Session, *DG.MessageCreate)

type Command struct {
	Name        string
	Constructor HandlerConstructor
}

var Commands = []Command{
	// Administration commands
	{"prefix", Prefix},
	{"setprefix", SetPrefix},
	{"addchan", AddChannel},
	{"rmvchan", RemoveChannel},
	{"addadmin", AddAdmin},
	{"rmvadmin", RemoveAdmin},
	{"admins", Admins},
	{"chanlist", Channels},
	{"info", Info},
	{"play", Play},
	{"reset", Reset},

	// Game commands
	{"spawn", Spawn},
	{"grab", Grab},
}
