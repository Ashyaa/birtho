package bot

import (
	"fmt"
	"strings"
	"time"

	DG "github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
)

const (
	FirstPageLabel    = "\u2060 \u2060 \u2060 First \u2060 \u2060 \u2060"
	PreviousPageLabel = "Previous"
	NextPageLabel     = "\u2060 \u2060 \u2060 Next \u2060 \u2060 \u2060"
	LastPageLabel     = "\u2060 \u2060 \u2060 Last \u2060 \u2060 \u2060"
)

type Listable interface {
	string
}

type Menu struct {
	cID      string   // Discord channel ID
	mID      string   // Discord message ID
	gID      string   // Discord guild ID
	L        []string // List to be displayed in a embed message
	Images   []string // List of images to be set as thumbnails on each page
	header   string
	footer   string
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
	if maxPage == 0 {
		maxPage = 1
	}
	res := Menu{
		L:       list,
		page:    1,
		expires: time.Now().Local().Add(time.Duration(60) * time.Second),
		maxPage: maxPage,
		size:    size,
		cID:     cID,
		gID:     gID,
		Images:  make([]string, maxPage),
	}
	return res
}

func (m *Menu) SetTitle(title string) {
	m.title = title
}

func (m *Menu) SetSubtitle(subtitle string) {
	m.subtitle = subtitle
}

func (m *Menu) SetImages(images []string) {
	nbImages := max(m.maxPage, len(images))
	for i := 0; i < nbImages; i++ {
		m.Images[i] = images[i]
	}
}

func (m *Menu) SetHeader(header string) {
	m.header = header
}

func (m *Menu) SetFooter(footer string) {
	m.footer = footer
}

func (m *Menu) GetComponents() *[]DG.MessageComponent {
	return &[]DG.MessageComponent{
		DG.ActionsRow{
			Components: []DG.MessageComponent{
				DG.Button{
					Label:    FirstPageLabel,
					CustomID: FirstPageLabel,
					Disabled: m.page == 1,
				},
				DG.Button{
					Label:    PreviousPageLabel,
					CustomID: PreviousPageLabel,
					Disabled: m.page == 1,
				},
				DG.Button{
					Label:    NextPageLabel,
					CustomID: NextPageLabel,
					Disabled: m.page == m.maxPage,
				},
				DG.Button{
					Label:    LastPageLabel,
					CustomID: LastPageLabel,
					Disabled: m.page == m.maxPage,
				},
			},
		},
	}
}

func (m *Menu) Send(s *DG.Session, i *DG.Interaction) error {
	msg, err := SendEmbed(s, i, m.cID, m.render(), *m.GetComponents())
	if err == nil {
		m.mID = msg.ID
	}
	return err
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
	text := ""
	if m.subtitle != "" {
		text = m.subtitle + "\n"
	}
	if len(m.L) != 0 {
		data := append([]string{m.header}, m.L[minIndex:maxIndex]...)
		text += "```\n" + strings.Join(data, "\n") + "\n```"
	}
	footer := fmt.Sprintf("Page %d/%d", m.page, m.maxPage)
	if m.footer != "" {
		footer = m.footer + "\n" + footer
	}
	return embed.NewEmbed().
		SetTitle(m.title).
		SetDescription(text).
		SetThumbnail(m.Images[m.page-1]).
		SetFooter(footer).
		SetColor(0x555555).MessageEmbed
}

func (m *Menu) PreviousPage(s *DG.Session) {
	m.page -= 1
	if m.page < 1 {
		m.page = 1
	}
	edit := DG.NewMessageEdit(m.cID, m.mID).SetEmbed(m.render())
	edit.Components = m.GetComponents()
	s.ChannelMessageEditComplex(edit)
}

func (m *Menu) NextPage(s *DG.Session) {
	m.page += 1
	if m.page > m.maxPage {
		m.page = m.maxPage
	}
	edit := DG.NewMessageEdit(m.cID, m.mID).SetEmbed(m.render())
	edit.Components = m.GetComponents()
	s.ChannelMessageEditComplex(edit)
}

func (m *Menu) FirstPage(s *DG.Session) {
	m.page = 1
	edit := DG.NewMessageEdit(m.cID, m.mID).SetEmbed(m.render())
	edit.Components = m.GetComponents()
	s.ChannelMessageEditComplex(edit)
}

func (m *Menu) LastPage(s *DG.Session) {
	m.page = m.maxPage
	edit := DG.NewMessageEdit(m.cID, m.mID).SetEmbed(m.render())
	edit.Components = m.GetComponents()
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
			if m, ok := b.Menus[ID]; ok {
				msg, err := b.s.ChannelMessage(b.Menus[ID].cID, b.Menus[ID].mID)
				if err != nil || len(msg.Embeds) == 0 {
					return
				}
				edit := DG.NewMessageEdit(m.cID, m.mID).SetEmbed(m.render())
				edit.Components = &[]DG.MessageComponent{}
				b.s.ChannelMessageEditComplex(edit)
				delete(b.Menus, ID)
			}
		}
		b.Info("purged menus, remaining menus: %d", len(b.Menus))
	}
}

func PageReact(b *Bot) func(*DG.Session, *DG.InteractionCreate) {
	return func(s *DG.Session, i *DG.InteractionCreate) {
		channel := i.ChannelID
		if i.Message == nil {
			b.Error("page interaction does not come from a button")
			s.InteractionRespond(i.Interaction, &DG.InteractionResponse{
				Type: DG.InteractionResponseUpdateMessage,
			})
			return
		}
		mID := i.Message.ID

		menu, ok := b.Menus[id(mID, channel, i.GuildID)]
		if !ok {
			b.Info("nok")
			return
		}
		if menu.mID != mID || menu.cID != channel || menu.gID != i.GuildID {
			b.Info("incorrect")
			return
		}
		switch name := i.MessageComponentData().CustomID; name {
		case FirstPageLabel:
			menu.FirstPage(s)
		case PreviousPageLabel:
			menu.PreviousPage(s)
		case NextPageLabel:
			menu.NextPage(s)
		case LastPageLabel:
			menu.LastPage(s)
		default:
			b.Warn("unknown button %s", name)
		}
		b.Menus[menu.ID()] = menu
		s.InteractionRespond(i.Interaction, &DG.InteractionResponse{
			Type: DG.InteractionResponseUpdateMessage,
		})
	}
}
