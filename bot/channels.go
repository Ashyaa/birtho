package bot

import (
	"fmt"
	"strings"

	"github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
)

func RemoveChannel(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
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
		if !serv.IsAdmin(m.Author.ID) {
			return
		}

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
