package bot

import (
	"fmt"
	"strings"

	"github.com/asdine/storm/v3"
	"github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
	LR "github.com/sirupsen/logrus"
)

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
	{"prefix", Prefix},
	{"setprefix", SetPrefix},
	{"addchan", AddChannel},
	{"rmvchan", RemoveChannel},
	{"chanlist", Channels},
}

func New(log LR.Logger) (*Bot, error) {
	conf, err := ReadConfig()
	if err != nil {
		log.Error("error reading config: ", err)
		return nil, err
	}

	res := Bot{Log: log}

	res.dg, err = DG.New("Bot " + conf.Token)
	if err != nil {
		log.Error("error creating session: ", err)
		return nil, err
	}

	// Open the database
	res.OpenDB()

	// Install command handlers
	res.SetupCommands()

	res.dg.Identify.Intents = DG.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = res.dg.Open()
	if err != nil {
		res.ErrorE(err, "error opening connection")
		return nil, err
	}

	return &res, nil
}

type MsgCreate func(s *DG.Session, m *DG.MessageCreate)

func (b *Bot) SetupCommands() {
	for _, cmd := range Commands {
		b.dg.AddHandler(cmd.Constructor(b, cmd.Name))
	}
}

func (b *Bot) Tag() string {
	if b.tag == "" {
		b.tag = "<@!" + b.dg.State.User.ID + ">"
	}
	return b.tag
}

// Returns true and the command content if the message triggers the command, else false and an empty string
func (b *Bot) triggered(s *DG.Session, m *DG.MessageCreate, prefix, command string) ([]string, bool) {
	tagCommand := b.Tag() + " " + command
	payload := []string{}
	fields := strings.Fields(m.Content)
	if strings.HasPrefix(m.Content, tagCommand) {
		if len(fields) > 2 {
			payload = fields[2:]
		}
		return payload, true
	}
	prefixCommand := prefix + command
	if strings.HasPrefix(m.Content, prefixCommand) {
		if len(fields) > 1 {
			payload = fields[1:]
		}
		return payload, true
	}
	return payload, false
}

func SetPrefix(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		serv := b.GetServer(m.GuildID)
		channel := m.ChannelID

		payload, ok := b.triggered(s, m, serv.Prefix, cmd)
		if !ok {
			return
		}
		b.Info("command %s triggered", cmd)

		if len(payload) == 0 {
			msg := fmt.Sprintf("usage: `%s%s <prefix>`", serv.Prefix, cmd)
			s.ChannelMessageSend(channel, msg)
		}

		serv.Prefix = payload[0]
		b.SaveServer(serv)

		msg := fmt.Sprintf("Bot prefix for this server was set to `%s`", serv.Prefix)
		s.ChannelMessageSend(channel, msg)
	}
}

func Prefix(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		serv := b.GetServer(m.GuildID)
		channel := m.ChannelID

		_, ok := b.triggered(s, m, serv.Prefix, cmd)
		if !ok {
			return
		}
		b.Info("command %s triggered", cmd)

		msg := fmt.Sprintf("Bot prefix for this server is `%s`", serv.Prefix)
		s.ChannelMessageSend(channel, msg)
	}
}

func RemoveChannel(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		serv := b.GetServer(m.GuildID)
		channel := m.ChannelID

		payload, ok := b.triggered(s, m, serv.Prefix, cmd)
		if !ok {
			return
		}

		if len(payload) == 0 {
			msg := fmt.Sprintf("No channel provided!\nusage: `%s%s <channel>`", serv.Prefix, cmd)
			s.ChannelMessageSend(channel, msg)
			return
		}

		targetChannel, ok := util.StripChannelTag(payload[0])
		if !ok || !util.IsValidChannel(s, m.GuildID, targetChannel) {
			msg := fmt.Sprintf("Channel `%s` is not a valid channel", payload[0])
			s.ChannelMessageSend(channel, msg)
			return
		}

		serv.Channels = util.Remove(serv.Channels, targetChannel)
		b.SaveServer(serv)

		msg := fmt.Sprintf("Removed channel %s from the list of spawn channels!", payload[0])
		s.ChannelMessageSend(channel, msg)
	}
}

func AddChannel(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		serv := b.GetServer(m.GuildID)
		channel := m.ChannelID

		payload, ok := b.triggered(s, m, serv.Prefix, cmd)
		if !ok {
			return
		}

		if len(payload) == 0 {
			msg := fmt.Sprintf("No channel provided!\nusage: `%s%s <channel>`", serv.Prefix, cmd)
			s.ChannelMessageSend(channel, msg)
			return
		}

		targetChannel, ok := util.StripChannelTag(payload[0])
		if !ok || !util.IsValidChannel(s, m.GuildID, targetChannel) {
			msg := fmt.Sprintf("Channel `%s` is not a valid channel", payload[0])
			s.ChannelMessageSend(channel, msg)
			return
		}

		serv.Channels = util.AppendUnique(serv.Channels, targetChannel)
		b.SaveServer(serv)

		msg := fmt.Sprintf("Added channel %s to the list of spawn channels!", payload[0])
		s.ChannelMessageSend(channel, msg)
	}
}

func Channels(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		serv := b.GetServer(m.GuildID)
		channel := m.ChannelID

		_, ok := b.triggered(s, m, serv.Prefix, cmd)
		if !ok {
			return
		}

		if len(serv.Channels) == 0 {
			msg := "No spawn channels setup!"
			s.ChannelMessageSend(channel, msg)
			return
		}

		channelTags := []string{}
		for _, cid := range serv.Channels {
			channelTags = append(channelTags, util.BuildChannelTag(cid))
		}

		msg := fmt.Sprintf("List of spawn channels: %s", strings.Join(channelTags, ", "))
		s.ChannelMessageSend(channel, msg)
	}
}

func (b *Bot) Stop() {
	b.Info("closing database")
	err := b.db.Close()
	if err != nil {
		b.ErrorE(err, "closing database")
	}
}
