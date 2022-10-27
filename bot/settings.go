package bot

import (
	"fmt"
	"strings"
	"time"

	U "github.com/ashyaa/birtho/util"
	embed "github.com/clinet/discordgo-embed"
)

func SetPrefix(b *Bot, p CommandParameters) {
	p.S.Prefix = p.Options["prefix"].(string)
	b.SaveServer(p.S)

	msg := fmt.Sprintf("Bot prefix for this server was set to `%s`", p.S.Prefix)
	SendText(b.s, p.I, p.CID, msg)
}

func Prefix(b *Bot, p CommandParameters) {
	msg := fmt.Sprintf("Bot prefix for this server is `%s`", p.S.Prefix)
	SendText(b.s, p.I, p.CID, msg)
}

func SetCooldown(b *Bot, p CommandParameters) {
	minNbSeconds := p.Options["minimum"].(int)
	maxNbSeconds := p.Options["maximum"].(int)
	minDelay := time.Duration(minNbSeconds) * time.Second
	maxDelay := time.Duration(maxNbSeconds) * time.Second
	if minNbSeconds < 0 {
		msg := fmt.Sprintf("The minimum delay `%v` cannot be negative.", minDelay)
		SendText(b.s, p.I, p.CID, msg)
		return
	}
	if minNbSeconds > maxNbSeconds {
		msg := fmt.Sprintf("Error: `%v` is superior to `%v`", minDelay, maxDelay)
		SendText(b.s, p.I, p.CID, msg)
		return
	}
	p.S.G.MinDelay = minDelay
	p.S.G.VariableDelay = maxNbSeconds - minNbSeconds + 1
	b.SaveServer(p.S)

	msg := fmt.Sprintf("Mininum cooldown set to `%s`.\n"+
		"Maxinum cooldown set to `%s`.\n", minDelay, maxDelay)
	SendText(b.s, p.I, p.CID, msg)
}

func SetStay(b *Bot, p CommandParameters) {
	nbSeconds := p.Options["duration"].(int)
	if nbSeconds <= 0 {
		msg := fmt.Sprintf("`%d` is not a valid number of seconds", nbSeconds)
		SendText(b.s, p.I, p.CID, msg)
		return
	}

	p.S.G.StayTime = time.Duration(nbSeconds) * time.Second
	b.SaveServer(p.S)

	msg := fmt.Sprintf("Stay time set to `%s`", p.S.G.StayTime)
	SendText(b.s, p.I, p.CID, msg)
}

func Info(b *Bot, p CommandParameters) {
	// Create new embed message
	msg := embed.NewEmbed().SetTitle("Server info").SetColor(0xaaddee)

	// Show play status
	status := "off"
	if p.S.G.On {
		status = "on"
	}
	msg.AddField("Play status", fmt.Sprintf("`%s`", status))

	// Show game status
	game := "`ongoing`"
	if p.S.G.Finished {
		game = fmt.Sprintf("`finished (winner: `%s`)`", U.BuildUserTag(p.S.G.Winner))
	}
	msg.AddField("Game status", game)

	if p.S.G.On {
		msg.AddField("Next spawn", U.Timestamp(p.S.G.NextSpawn))
	}

	// Show configured cooldown
	maxDelay := p.S.G.MinDelay + time.Duration(p.S.G.VariableDelay-1)*time.Second
	msg.AddField("Cooldown", fmt.Sprintf("`%v - %v`", p.S.G.MinDelay, maxDelay))

	// Show configured monster stay time
	msg.AddField("Monster stay time", fmt.Sprintf("`%v`", p.S.G.StayTime))

	// Show configured prefix
	msg.AddField("Prefix", fmt.Sprintf("`%s`", p.S.Prefix))

	// Show the configured list of admins
	userTags := []string{}
	for _, uid := range p.S.Admins {
		userTags = append(userTags, U.BuildUserTag(uid))
	}
	admins := "None"
	if len(p.S.Admins) != 0 {
		admins = strings.Join(userTags, ", ")
	}
	msg.AddField("Admins", admins)

	// Show the configured list of spawn channels
	channelTags := []string{}
	for _, cid := range p.S.Channels {
		channelTags = append(channelTags, U.BuildChannelTag(cid))
	}
	channels := "None"
	if len(p.S.Channels) != 0 {
		channels = strings.Join(channelTags, ", ")
	}
	msg.AddField("Channels", channels)
	SendEmbed(b.s, p.I, p.CID, msg.MessageEmbed)
}
