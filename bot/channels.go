package bot

import (
	"fmt"
	"strings"

	U "github.com/ashyaa/birtho/util"
)

func RemoveChannel(b *Bot, p CommandParameters) {
	targetChannel := p.Options["channel"].(string)
	tag := U.BuildChannelTag(targetChannel)
	if !U.IsValidChannel(b.s, p.GID, targetChannel) {
		msg := fmt.Sprintf("Channel `%s` is not a valid channel", tag)
		SendText(b.s, p.I, p.CID, msg)
		return
	}

	p.S.Channels = U.Remove(p.S.Channels, targetChannel)
	b.SaveServer(p.S)

	msg := fmt.Sprintf("Removed channel %s from the list of spawn channels!", tag)
	SendText(b.s, p.I, p.CID, msg)
}

func AddChannel(b *Bot, p CommandParameters) {
	targetChannel := p.Options["channel"].(string)
	tag := U.BuildChannelTag(targetChannel)
	if !U.IsValidChannel(b.s, p.GID, targetChannel) {
		msg := fmt.Sprintf("Channel `%s` is not a valid channel", tag)
		SendText(b.s, p.I, p.CID, msg)
		return
	}

	p.S.Channels = U.AppendUnique(p.S.Channels, targetChannel)
	b.SaveServer(p.S)

	msg := fmt.Sprintf("Added channel %s to the list of spawn channels!", tag)
	SendText(b.s, p.I, p.CID, msg)
}

func Channels(b *Bot, p CommandParameters) {
	if len(p.S.Channels) == 0 {
		msg := "No spawn channels setup!"
		SendText(b.s, p.I, p.CID, msg)
		return
	}

	channelTags := []string{}
	for _, cid := range p.S.Channels {
		channelTags = append(channelTags, U.BuildChannelTag(cid))
	}

	msg := fmt.Sprintf("List of spawn channels: %s", strings.Join(channelTags, ", "))
	SendText(b.s, p.I, p.CID, msg)
}
