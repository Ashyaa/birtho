package bot

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	U "github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
)

func Play(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
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
			msg := fmt.Sprintf("usage: `%s%s <on|off>`", serv.Prefix, cmd)
			s.ChannelMessageSend(channel, msg)
			return
		}
		arg := strings.ToLower(payload[0])
		if arg != "on" && arg != "off" {
			msg := fmt.Sprintf("usage: `%s%s <on|off>`", serv.Prefix, cmd)
			s.ChannelMessageSend(channel, msg)
			return
		}

		serv.G.On = arg == "on"
		b.SaveServer(serv)

		status := "off"
		if serv.G.On {
			status = "on"
		}

		msg := fmt.Sprintf("Play status set to `%s`", status)
		s.ChannelMessageSend(channel, msg)
	}
}

func Spawn(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		serv := b.GetServer(m.GuildID)
		if !serv.IsAdmin(m.Author.ID) {
			return
		}

		channel := m.ChannelID
		if !U.Contains(serv.Channels, channel) {
			return
		}

		_, ok := b.triggered(s, m, serv.Prefix, cmd)
		if !ok {
			return
		}
		b.Info("command %s triggered", cmd)

		spawn := Item{ID: strconv.Itoa(rand.Intn(100))}

		text := fmt.Sprintf("Spawn ID: `%s`", spawn.ID)
		msg, err := s.ChannelMessageSend(channel, text)
		if err != nil {
			b.ErrorE(err, "spawn message")
			return
		}
		spawn.Message = msg.ID
		serv.G.Items[channel] = spawn
		b.SaveServer(serv)

		time.AfterFunc(5*time.Second, func() {
			curServ := b.GetServer(m.GuildID)
			_, ok := curServ.G.Items[channel]
			if !ok {
				return
			}
			delete(curServ.G.Items, channel)
			b.SaveServer(curServ)

			s.ChannelMessageEdit(channel, msg.ID, "It's dead Jim")
		})
	}
}

func Grab(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		user := m.Author.ID
		if user == s.State.User.ID {
			return
		}

		serv := b.GetServer(m.GuildID)

		channel := m.ChannelID
		if !U.Contains(serv.Channels, channel) {
			return
		}

		_, ok := b.triggered(s, m, serv.Prefix, cmd)
		if !ok {
			return
		}
		spawn, ok := serv.G.Items[channel]
		if !ok {
			return
		}
		b.Info("command %s triggered", cmd)

		text := fmt.Sprintf("%s grabbed item `%s`", U.BuildUserTag(user), spawn.ID)
		s.ChannelMessageEdit(channel, spawn.Message, text)

		_, ok = serv.Users[user]
		if !ok {
			serv.Users[user] = make([]string, 0)
		}
		serv.Users[user] = U.AppendUnique(serv.Users[user], spawn.ID)
		delete(serv.G.Items, channel)

		b.SaveServer(serv)
	}
}
