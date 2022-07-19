package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	U "github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
)

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

func SetCooldown(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
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
		if len(payload) <= 1 {
			msg := fmt.Sprintf("No enough arguments! \nUsage: `%s%s <minCD> <maxCD>`\n"+
				"Both cooldowns are expected to be a time in seconds.", serv.Prefix, cmd)
			s.ChannelMessageSend(channel, msg)
			return
		}

		b.Info("command %s triggered", cmd)
		minNbSeconds, err := strconv.Atoi(payload[0])
		if err != nil {
			msg := fmt.Sprintf("`%s` is not a valid number of seconds", payload[0])
			s.ChannelMessageSend(channel, msg)
			return
		}
		maxNbSeconds, err := strconv.Atoi(payload[1])
		if err != nil {
			msg := fmt.Sprintf("`%s` is not a valid number of seconds", payload[1])
			s.ChannelMessageSend(channel, msg)
			return
		}
		minDelay := time.Duration(minNbSeconds) * time.Second
		if minNbSeconds < 0 {
			msg := fmt.Sprintf("The minimum delay `%v` cannot be negative.", minDelay)
			s.ChannelMessageSend(channel, msg)
			return
		}
		maxDelay := time.Duration(maxNbSeconds) * time.Second
		if minNbSeconds > maxNbSeconds {
			msg := fmt.Sprintf("Error: `%v` is superior to `%v`", minDelay, maxDelay)
			s.ChannelMessageSend(channel, msg)
			return
		}
		serv.G.MinDelay = minDelay
		serv.G.VariableDelay = maxNbSeconds - minNbSeconds + 1
		b.SaveServer(serv)

		msg := fmt.Sprintf("Mininum cooldown set to `%s`.\n"+
			"Maxinum cooldown set to `%s`.\n", minDelay, maxDelay)
		s.ChannelMessageSend(channel, msg)
	}
}

func SetStay(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
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
		if len(payload) == 0 {
			msg := fmt.Sprintf("No enough arguments! \nUsage: `%s%s <delay>`\n"+
				"The stay time is expected to be a time in seconds.", serv.Prefix, cmd)
			s.ChannelMessageSend(channel, msg)
			return
		}

		b.Info("command %s triggered", cmd)
		nbSeconds, err := strconv.Atoi(payload[0])
		if err != nil || nbSeconds <= 0 {
			msg := fmt.Sprintf("`%s` is not a valid number of seconds", payload[0])
			s.ChannelMessageSend(channel, msg)
			return
		}

		serv.G.StayTime = time.Duration(nbSeconds) * time.Second
		b.SaveServer(serv)

		msg := fmt.Sprintf("Stay time set to `%s`", serv.G.StayTime)
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

		// Show configured cooldown
		maxDelay := serv.G.MinDelay + time.Duration(serv.G.VariableDelay-1)*time.Second
		msg.AddField("Cooldown", fmt.Sprintf("`%v - %v`", serv.G.MinDelay, maxDelay))

		// Show configured monster stay time
		msg.AddField("Monster stay time", fmt.Sprintf("`%v`", serv.G.StayTime))

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
