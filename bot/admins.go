package bot

import (
	"fmt"
	"strings"

	U "github.com/ashyaa/birtho/util"
)

func RemoveAdmin(b *Bot, p CommandParameters) {
	user := p.Options["user"].(string)
	if !U.IsUserInServer(b.s, p.GID, user) {
		msg := fmt.Sprintf("User `%s` is not a valid user", user)
		SendText(b.s, p.I, p.CID, msg)
		return
	}

	p.S.Admins = U.Remove(p.S.Admins, user)
	b.SaveServer(p.S)

	msg := fmt.Sprintf("Removed user %s from the list of bot admins!", U.BuildUserTag(user))
	SendText(b.s, p.I, p.CID, msg)
}

func AddAdmin(b *Bot, p CommandParameters) {
	user := p.Options["user"].(string)
	if !U.IsUserInServer(b.s, p.GID, user) {
		msg := fmt.Sprintf("User `%s` is not a valid user", user)
		SendText(b.s, p.I, p.CID, msg)
		return
	}

	p.S.Admins = U.AppendUnique(p.S.Admins, user)
	b.SaveServer(p.S)

	msg := fmt.Sprintf("Added user %s to the list of bot admins!", U.BuildUserTag(user))
	SendText(b.s, p.I, p.CID, msg)
}

func Admins(b *Bot, p CommandParameters) {
	if len(p.S.Admins) == 0 {
		msg := "No admins set: everyone can use the configuration commands"
		SendText(b.s, p.I, p.CID, msg)
		return
	}

	userTags := []string{}
	for _, uid := range p.S.Admins {
		userTags = append(userTags, U.BuildUserTag(uid))
	}

	msg := fmt.Sprintf("List of bot admins: %s", strings.Join(userTags, ", "))
	SendText(b.s, p.I, p.CID, msg)
}
