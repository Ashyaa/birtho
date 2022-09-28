package bot

import (
	"fmt"
	"strings"
	"time"

	DG "github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
)

type Listable interface {
	string
}

type Menu struct {
	cID      string   // Discord channel ID
	mID      string   // Discord message ID
	gID      string   // Discord guild ID
	L        []string // List to be displayed in a embed message
	header   string
	expires  time.Time
	page     int // Current page
	maxPage  int
	size     int // Number of elements per page
	title    string
	subtitle string
}

func NewMenu(list []string, size int, cID, gID string) Menu {
	maxPage := len(list) / size
	if len(list)%size > 0 {
		maxPage += 1
	}
	res := Menu{
		L:       list,
		page:    1,
		expires: time.Now().Local().Add(time.Duration(60) * time.Second),
		maxPage: maxPage,
		size:    size,
		cID:     cID,
		gID:     gID,
	}
	return res
}

func (m *Menu) SetTitle(title string) {
	m.title = title
}

func (m *Menu) SetSubtitle(subtitle string) {
	m.subtitle = subtitle
}

func (m *Menu) SetHeader(header string) {
	m.header = header
}

func (m *Menu) Send(s *DG.Session, i *DG.Interaction) error {
	msg, err := SendEmbed(s, i, m.cID, m.render())
	if err != nil {
		return err
	}
	m.mID = msg.ID
	err = s.MessageReactionAdd(m.cID, m.mID, "⬅️")
	if err != nil {
		return err
	}
	err = s.MessageReactionAdd(m.cID, m.mID, "➡️")
	if err != nil {
		return err
	}
	return nil
}

func id(mID, cID, gID string) string {
	return mID + "." + cID + "." + gID
}

func (m *Menu) ID() string {
	return id(m.mID, m.cID, m.gID)
}

func (m *Menu) render() *DG.MessageEmbed {
	minIndex := (m.page - 1) * m.size
	if minIndex < 0 {
		minIndex = 0
	}
	maxIndex := (m.page) * m.size
	if maxIndex > len(m.L) {
		maxIndex = len(m.L)
	}
	data := append([]string{m.header}, m.L[minIndex:maxIndex]...)
	text := "```\n" + strings.Join(data, "\n") + "\n```"
	if m.subtitle != "" {
		text = m.subtitle + "\n" + text
	}
	return embed.NewEmbed().
		SetTitle(m.title).
		SetDescription(text).
		SetFooter(fmt.Sprintf("Page %d/%d", m.page, m.maxPage)).
		SetColor(0x555555).MessageEmbed
}

func (m *Menu) PreviousPage(s *DG.Session) {
	m.page -= 1
	if m.page < 1 {
		m.page = 1
	}
	edit := DG.NewMessageEdit(m.cID, m.mID).SetEmbed(m.render())
	s.ChannelMessageEditComplex(edit)
}

func (m *Menu) NextPage(s *DG.Session) {
	m.page += 1
	if m.page > m.maxPage {
		m.page = m.maxPage
	}
	edit := DG.NewMessageEdit(m.cID, m.mID).SetEmbed(m.render())
	s.ChannelMessageEditComplex(edit)
}

func purgeMenus(b *Bot) func() {
	return func() {
		toRemove := []string{}
		now := time.Now().Local()
		for ID, menu := range b.Menus {
			if now.After(menu.expires) {
				toRemove = append(toRemove, ID)
			}
		}
		for _, ID := range toRemove {
			if _, ok := b.Menus[ID]; ok {
				b.s.MessageReactionsRemoveAll(b.Menus[ID].cID, b.Menus[ID].mID)
				delete(b.Menus, ID)
			}
		}
		b.Info("purged menus, remaining menus: %d", len(b.Menus))
	}
}

func PageReact(b *Bot) func(*DG.Session, *DG.MessageReactionAdd) {
	return func(s *DG.Session, ra *DG.MessageReactionAdd) {
		if ra.UserID == b.UserID {
			return
		}
		channel := ra.ChannelID
		mID := ra.MessageID

		menu, ok := b.Menus[id(mID, channel, ra.GuildID)]
		if !ok {
			return
		}
		if menu.mID != mID || menu.cID != channel || menu.gID != ra.GuildID {
			return
		}

		if ra.Emoji.Name == "⬅️" {
			menu.PreviousPage(s)
			b.Info("previous page")
			s.MessageReactionRemove(channel, mID, "⬅️", ra.UserID)
		} else if ra.Emoji.Name == "➡️" {
			menu.NextPage(s)
			b.Info("next page")
			s.MessageReactionRemove(channel, mID, "➡️", ra.UserID)
		}
		b.Menus[menu.ID()] = menu
	}
}
