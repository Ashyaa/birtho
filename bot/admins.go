package bot

import (
	"fmt"
	"strings"

	U "github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
)

func RemoveAdmin(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
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

		if len(payload) == 0 {
			msg := fmt.Sprintf("No channel provided!\nusage: `%s%s <channel>`", serv.Prefix, cmd)
			s.ChannelMessageSend(channel, msg)
			return
		}

		user, ok := U.StripUserTag(payload[0])
		if !ok {
			msg := fmt.Sprintf("User `%s` is not a valid user", payload[0])
			s.ChannelMessageSend(channel, msg)
			return
		}

		serv.Admins = U.Remove(serv.Admins, user)
		b.SaveServer(serv)

		msg := fmt.Sprintf("Removed user %s from the list of bot admins!", payload[0])
		s.ChannelMessageSend(channel, msg)
	}
}

func AddAdmin(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
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

		if len(payload) == 0 {
			msg := fmt.Sprintf("No user provided!\nusage: `%s%s <user>`", serv.Prefix, cmd)
			s.ChannelMessageSend(channel, msg)
			return
		}

		user, ok := U.StripUserTag(payload[0])
		if !ok || !U.IsUserInServer(s, m.GuildID, user) {
			msg := fmt.Sprintf("User `%s` is not a valid user", payload[0])
			s.ChannelMessageSend(channel, msg)
			return
		}

		serv.Admins = U.AppendUnique(serv.Admins, user)
		b.SaveServer(serv)

		msg := fmt.Sprintf("Added user %s to the list of bot admins!", payload[0])
		s.ChannelMessageSend(channel, msg)
	}
}

func Admins(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
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

		if len(serv.Admins) == 0 {
			msg := "No admins set: everyone can use the configuration commands"
			s.ChannelMessageSend(channel, msg)
			return
		}

		userTags := []string{}
		for _, uid := range serv.Admins {
			userTags = append(userTags, U.BuildUserTag(uid))
		}

		msg := fmt.Sprintf("List of bot admins: %s", strings.Join(userTags, ", "))
		s.ChannelMessageSend(channel, msg)
	}
}
