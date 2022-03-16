package bot

import (
	"github.com/asdine/storm/v3"
	U "github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
	LR "github.com/sirupsen/logrus"
)

const DefaultPrefix = "b!"

type Item struct {
	ID      string
	Message string
}

type Game struct {
	On        bool
	Items     map[string]Item
	messageID string
}

type Server struct {
	ID       string
	Prefix   string
	G        Game
	Channels []string
	Admins   []string
	Users    map[string][]string
}

func (s Server) IsAdmin(uid string) bool {
	// Everyone is an admin when no admins are registered
	if len(s.Admins) == 0 {
		return true
	}
	return U.Contains(s.Admins, uid)
}

type Bot struct {
	dg  *DG.Session
	db  *storm.DB
	Log LR.Logger
	tag string
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

	// Game commands
	{"spawn", Spawn},
	{"grab", Grab},
}
