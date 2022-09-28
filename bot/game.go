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

func Play(b *Bot, p CommandParameters) {
	arg := strings.ToLower(p.Options["state"].(string))
	if arg != "on" && arg != "off" {
		msg := fmt.Sprintf("usage: `%s%s <on|off>`", p.S.Prefix, p.Name)
		SendText(b.s, p.I, p.CID, msg)
		return
	}

	p.S.Cooldown()
	p.S.G.On = arg == "on"
	b.SaveServer(p.S)

	status := "off"
	if p.S.G.On {
		status = "on"
	}

	msg := fmt.Sprintf("Play status set to `%s`", status)
	SendText(b.s, p.I, p.CID, msg)
}

func Reset(b *Bot, p CommandParameters) {
	p.S.Users = make(map[string][]string)
	b.SaveServer(p.S)

	msg := "Cleared all players' item list"
	SendText(b.s, p.I, p.CID, msg)
}

func Spawn(b *Bot, p CommandParameters) {
	isManualCommand := p.S.IsAdmin(p.UID) && p.IsUserTriggered
	if !p.S.CanSpawn(p.CID) && !isManualCommand {
		return
	}

	monster := b.RandomMonster()
	spawn := MonsterSpawn{
		ID: strconv.Itoa(monster.ID),
	}

	msg, err := SendEmbed(b.s, p.I, p.CID, embed.NewEmbed().
		SetTitle("A visitor has come!").
		SetDescription(fmt.Sprintf("**%s** appeared!", monster.Name)).
		SetColor(0x00FF00).
		SetImage(monster.URL).MessageEmbed)
	if err != nil {
		b.ErrorE(err, "spawn message")
		return
	}
	spawn.Message = msg.ID
	p.S.G.Monsters[p.CID] = spawn
	p.S.Cooldown()
	b.SaveServer(p.S)

	time.AfterFunc(p.S.G.StayTime, func() {
		curServ := b.GetServer(p.GID)
		_, ok := curServ.G.Monsters[p.CID]
		if !ok {
			return
		}
		delete(curServ.G.Monsters, p.CID)
		b.SaveServer(curServ)

		edit := DG.NewMessageEdit(p.CID, msg.ID).SetEmbed(embed.NewEmbed().
			SetTitle("The visitor has left.").
			SetDescription(fmt.Sprintf("**%s** left...", monster.Name)).
			SetColor(0xFF0000).MessageEmbed)
		b.s.ChannelMessageEditComplex(edit)
	})
}

func Grab(b *Bot, p CommandParameters) {
	channel := p.CID
	if !U.Contains(p.S.Channels, channel) {
		return
	}

	spawn, ok := p.S.G.Monsters[channel]
	if !ok {
		return
	}

	if _, ok = p.S.Users[p.UID]; !ok {
		p.S.Users[p.UID] = make([]string, 0)
	}

	if trickOrTreat() {
		monster := b.Monsters[spawn.ID]
		item := monster.RandomItem(b.Log)
		text := fmt.Sprintf("As a thank you for your kindness, **%s** gives %s one **%s**",
			monster.Name, U.BuildUserTag(p.UID), item.Name)
		b.s.ChannelMessageEditEmbed(channel, spawn.Message, embed.NewEmbed().
			SetTitle("The visitor has been pleased!").
			SetDescription(text).
			SetColor(0xFFFFFF).
			SetFooter(itemDescription(item)).
			SetImage(monster.URL).MessageEmbed)
		p.S.Users[p.UID] = U.AppendUnique(p.S.Users[p.UID], item.ID)
	} else {
		monster := b.Monsters[spawn.ID]
		text := fmt.Sprintf("%s scared **%s** away...", U.BuildUserTag(p.UID), monster.Name)
		b.s.ChannelMessageEditEmbed(channel, spawn.Message, embed.NewEmbed().
			SetTitle("The visitor has fled!").
			SetDescription(text).
			SetColor(0xFF0000).MessageEmbed)
	}
	delete(p.S.G.Monsters, channel)
	b.SaveServer(p.S)
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

func GiveRandom(b *Bot, p CommandParameters) {
	items := []string{}
	for item := range b.Items {
		items = append(items, item)
	}

	item := items[rand.Intn(len(items))]
	if _, ok := p.S.Users[p.UID]; !ok {
		p.S.Users[p.UID] = make([]string, 0)
	}
	p.S.Users[p.UID] = U.AppendUnique(p.S.Users[p.UID], item)
	b.SaveServer(p.S)

	msg := fmt.Sprintf("Gave you one `%s`", b.Items[item].Name)
	SendText(b.s, p.I, p.CID, msg)
}
