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

func (b *Bot) RandomMonster() Monster {
	if b.EqualMonsterChances {
		index := rand.Intn(len(b.MonsterIds))
		key := b.MonsterIds[index]
		return b.Monsters[key]
	}
	number := rand.Intn(10000) + 1
	for _, m := range b.Monsters {
		if m.Range.Belongs(number) {
			return m
		}
	}
	b.Fatal("invalid monster roll")
	return Monster{}
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

func Reset(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
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

		serv.Users = make(map[string][]string)
		b.SaveServer(serv)

		msg := "Cleared all players' item list"
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

		monster := b.RandomMonster()
		spawn := MonsterSpawn{
			ID: strconv.Itoa(monster.ID),
		}

		msg, err := s.ChannelMessageSendEmbed(channel, embed.NewEmbed().
			SetTitle("A visitor has come!").
			SetDescription(fmt.Sprintf("**%s** appeared!", monster.Name)).
			SetColor(0x00FF00).
			SetImage(monster.URL).MessageEmbed)
		if err != nil {
			b.ErrorE(err, "spawn message")
			return
		}
		spawn.Message = msg.ID
		serv.G.Monsters[channel] = spawn
		serv.Cooldown()
		b.SaveServer(serv)

		time.AfterFunc(serv.G.StayTime, func() {
			curServ := b.GetServer(m.GuildID)
			_, ok := curServ.G.Monsters[channel]
			if !ok {
				return
			}
			delete(curServ.G.Monsters, channel)
			b.SaveServer(curServ)

			edit := DG.NewMessageEdit(channel, msg.ID).SetEmbed(embed.NewEmbed().
				SetTitle("The visitor has left.").
				SetDescription(fmt.Sprintf("**%s** left...", monster.Name)).
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
		spawn, ok := serv.G.Monsters[channel]
		if !ok {
			return
		}
		b.Info("command %s triggered", cmd)

		if _, ok = serv.Users[user]; !ok {
			serv.Users[user] = make([]string, 0)
		}

		if trickOrTreat() {
			monster := b.Monsters[spawn.ID]
			item := monster.RandomItem(b.Log)
			text := fmt.Sprintf("As a thank you for your kindness, **%s** gives %s one **%s**",
				monster.Name, U.BuildUserTag(user), item.Name)
			s.ChannelMessageEditEmbed(channel, spawn.Message, embed.NewEmbed().
				SetTitle("The visitor has been pleased!").
				SetDescription(text).
				SetColor(0xFFFFFF).
				SetFooter(itemDescription(item)).
				SetImage(monster.URL).MessageEmbed)
			serv.Users[user] = U.AppendUnique(serv.Users[user], item.ID)
		} else {
			monster := b.Monsters[spawn.ID]
			text := fmt.Sprintf("%s scared **%s** away...", U.BuildUserTag(user), monster.Name)
			s.ChannelMessageEditEmbed(channel, spawn.Message, embed.NewEmbed().
				SetTitle("The visitor has fled!").
				SetDescription(text).
				SetColor(0xFF0000).MessageEmbed)
		}
		delete(serv.G.Monsters, channel)
		b.SaveServer(serv)
	}
}

func itemDescription(item Item) string {
	text := ""
	if item.Chance < 20 {
		text = "This item is rare. It must be worth a lot."
	} else if item.Chance < 50 {
		text = "This item is uncommon. You wonder where they got it..."
	} else {
		text = "This item is common. There's nothing special about it."
	}
	return text + " It has been added to your inventory."
}

func trickOrTreat() bool {
	n := rand.Intn(1000)
	return n > 499
}

func GiveRandom(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		usr := m.Author.ID
		if usr == b.UserID {
			return
		}

		serv := b.GetServer(m.GuildID)
		_, ok := b.triggered(s, m, serv.Prefix, cmd)
		if !ok || !serv.IsAdmin(usr) {
			return
		}
		b.Info("command %s triggered", cmd)

		items := []string{}
		for item := range b.Items {
			items = append(items, item)
		}

		item := items[rand.Intn(len(items))]
		if _, ok := serv.Users[usr]; !ok {
			serv.Users[usr] = make([]string, 0)
		}
		serv.Users[usr] = U.AppendUnique(serv.Users[usr], item)
		b.SaveServer(serv)

		msg := fmt.Sprintf("Gave you one `%s`", b.Items[item].Name)
		s.ChannelMessageSend(m.ChannelID, msg)
	}
}
