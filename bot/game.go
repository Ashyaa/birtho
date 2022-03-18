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
)

func (b *Bot) RandomItem() Item {
	index := rand.Intn(len(b.ItemIds))
	key := b.ItemIds[index]
	return b.Items[key]
}

func Play(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
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

		serv.Cooldown()
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
		if m.Author.ID == b.UserID {
			return
		}

		serv := b.GetServer(m.GuildID)
		_, ok := b.triggered(s, m, serv.Prefix, cmd)
		isManualCommand := serv.IsAdmin(m.Author.ID) && ok

		channel := m.ChannelID
		if !serv.CanSpawn(channel) && !isManualCommand {
			return
		}
		b.Info("command %s triggered", cmd)

		item := b.RandomItem()
		spawn := ItemSpawn{
			ID: strconv.Itoa(item.ID),
		}

		msg, err := s.ChannelMessageSendEmbed(channel, embed.NewEmbed().
			SetDescription(fmt.Sprintf("`%s` appeared!", item.Name)).
			SetColor(0x00FF00).
			SetImage(item.URL).MessageEmbed)
		if err != nil {
			b.ErrorE(err, "spawn message")
			return
		}
		spawn.Message = msg.ID
		serv.G.Items[channel] = spawn
		serv.Cooldown()
		b.SaveServer(serv)

		time.AfterFunc(5*time.Second, func() {
			curServ := b.GetServer(m.GuildID)
			_, ok := curServ.G.Items[channel]
			if !ok {
				return
			}
			delete(curServ.G.Items, channel)
			b.SaveServer(curServ)

			edit := DG.NewMessageEdit(channel, msg.ID).SetEmbed(embed.NewEmbed().
				SetDescription(fmt.Sprintf("`%s` left...", item.Name)).
				SetColor(0xFF0000).MessageEmbed)
			s.ChannelMessageEditComplex(edit)
		})
	}
}

func Grab(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		user := m.Author.ID
		if user == b.UserID {
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

		item := b.Items[spawn.ID]
		text := fmt.Sprintf("%s grabbed `%s`!", U.BuildUserTag(user), item.Name)
		s.ChannelMessageEditEmbed(channel, spawn.Message, embed.NewEmbed().
			SetDescription(text).
			SetColor(0xFFFFFF).
			SetImage(item.URL).MessageEmbed)

		_, ok = serv.Users[user]
		if !ok {
			serv.Users[user] = make([]string, 0)
		}
		serv.Users[user] = U.AppendUnique(serv.Users[user], spawn.ID)
		delete(serv.G.Items, channel)

		b.SaveServer(serv)
	}
}