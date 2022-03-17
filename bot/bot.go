package bot

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	U "github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
	LR "github.com/sirupsen/logrus"
)

func New(log LR.Logger) (*Bot, error) {
	conf, err := ReadConfig(log)
	if err != nil {
		log.Error("error reading config: ", err)
		return nil, err
	}

	res := Bot{Log: log}
	res.buildItemDict(conf)

	res.dg, err = DG.New("Bot " + conf.Token)
	if err != nil {
		log.Error("error creating session: ", err)
		return nil, err
	}

	// Open the database
	res.OpenDB()

	// Install command handlers
	res.SetupCommands()

	// Provide a seed for the pseudo-random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	res.dg.Identify.Intents = DG.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = res.dg.Open()
	if err != nil {
		res.ErrorE(err, "error opening connection")
		return nil, err
	}
	res.UserID = res.dg.State.User.ID
	res.Mention = U.BuildUserTag(res.UserID)

	return &res, nil
}

type MsgCreate func(s *DG.Session, m *DG.MessageCreate)

func (b *Bot) SetupCommands() {
	for _, cmd := range Commands {
		b.dg.AddHandler(cmd.Constructor(b, cmd.Name))
	}
}

func (b *Bot) Stop() {
	b.Info("closing database")
	err := b.db.Close()
	if err != nil {
		b.ErrorE(err, "closing database")
	}
}

// Returns true and the command content if the message triggers the command, else false and an empty string
func (b *Bot) triggered(s *DG.Session, m *DG.MessageCreate, prefix, command string) ([]string, bool) {
	tagCommand := b.Mention + " " + command
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

func (b *Bot) buildItemDict(conf Config) {
	b.ItemIds = make([]string, 0)
	b.Items = make(map[string]Item)
	for _, item := range conf.Items {
		if item.URL == "" {
			continue
		}
		key := strconv.Itoa(item.ID)
		b.ItemIds = append(b.ItemIds, key)
		b.Items[key] = item
	}
}

func SetPrefix(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == b.UserID {
			return
		}

		serv := b.GetServer(m.GuildID)
		if !serv.IsAdmin(m.Author.ID) {
			return
		}

		channel := m.ChannelID

		payload, ok := b.triggered(s, m, serv.Prefix, cmd)
		if !ok {
			return
		}
		b.Info("command %s triggered", cmd)

		if len(payload) == 0 {
			msg := fmt.Sprintf("usage: `%s%s <prefix>`", serv.Prefix, cmd)
			s.ChannelMessageSend(channel, msg)
			return
		}

		serv.Prefix = payload[0]
		b.SaveServer(serv)

		msg := fmt.Sprintf("Bot prefix for this server was set to `%s`", serv.Prefix)
		s.ChannelMessageSend(channel, msg)
	}
}

func Prefix(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == b.UserID {
			return
		}

		serv := b.GetServer(m.GuildID)
		if !serv.IsAdmin(m.Author.ID) {
			return
		}

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

func Info(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == b.UserID {
			return
		}

		serv := b.GetServer(m.GuildID)
		if !serv.IsAdmin(m.Author.ID) {
			return
		}

		channel := m.ChannelID

		_, ok := b.triggered(s, m, serv.Prefix, cmd)
		if !ok {
			return
		}
		b.Info("command %s triggered", cmd)

		// Create new embed message
		msg := embed.NewEmbed().SetTitle("Server info").SetColor(0xaaddee)

		// Show play status
		status := "off"
		if serv.G.On {
			status = "on"
		}
		msg.AddField("Play status", fmt.Sprintf("`%s`", status))

		if serv.G.On {
			msg.AddField("Next spawn", U.Timestamp(serv.G.NextSpawn))
		}

		// Show configured prefix
		msg.AddField("Prefix", fmt.Sprintf("`%s`", serv.Prefix))

		// Show the configured list of admins
		userTags := []string{}
		for _, uid := range serv.Admins {
			userTags = append(userTags, U.BuildUserTag(uid))
		}
		admins := "None"
		if len(serv.Admins) != 0 {
			admins = strings.Join(userTags, ", ")
		}
		msg.AddField("Admins", admins)

		// Show the configured list of spawn channels
		channelTags := []string{}
		for _, cid := range serv.Channels {
			channelTags = append(channelTags, U.BuildChannelTag(cid))
		}
		channels := "None"
		if len(serv.Channels) != 0 {
			channels = strings.Join(channelTags, ", ")
		}
		msg.AddField("Channels", channels)
		s.ChannelMessageSendEmbed(channel, msg.MessageEmbed)
	}
}
