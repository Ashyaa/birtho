package bot

import (
	"fmt"
	"sort"
	"time"

	DG "github.com/bwmarrin/discordgo"
)

func (b *Bot) GetUserScore(user string, serv Server) int {
	res := 0
	if _, ok := serv.Users[user]; !ok {
		serv.Users[user] = make([]string, 0)
	}
	for _, itemName := range serv.Users[user] {
		if item, ok := b.Items[itemName]; ok {
			res += item.Points
		}
	}
	return res
}

type scoreboard struct {
	userID string
	user   string
	score  int
	rank   string
}

type leaderboard []scoreboard

func (l leaderboard) Strings() []string {
	res := []string{}
	rankPlaces := len(fmt.Sprintf("%d", len(l))) + 2 // + suffix 'st', 'nd', 'rd', 'th'
	if rankPlaces < 4 {
		rankPlaces = 4
	}
	header := "Rank     Points     User"
	if rankPlaces > 4 {
		header = padLeft(header, rankPlaces-4)
	}
	res = append(res, header)
	if len(l) == 0 {
		return res
	}
	for _, sb := range l {
		rank := sb.rank
		rank = padLeft(rank, rankPlaces-len(rank))

		score := fmt.Sprintf("%d", sb.score)
		score = padLeft(score, 11-len(score))

		line := fmt.Sprintf("%s%s     %s", rank, score, sb.user)
		res = append(res, line)
	}
	return res
}

func rankString(n int) string {
	res := fmt.Sprintf("%d", n)
	if (n/10)%10 == 1 {
		return res + "th"
	}
	switch n % 10 {
	case 1:
		res += "st"
	case 2:
		res += "nd"
	case 3:
		res += "rd"
	default:
		res += "th"
	}
	return res
}

func padLeft(s string, padding int) string {
	if padding <= 0 {
		return s
	}
	res := "⁠" + s

	for i := 0; i < padding; i++ {
		res = " " + res
	}
	return res
}

func (b *Bot) getLeaderBoard(serv Server, s *DG.Session) leaderboard {
	lb := leaderboard{}
	for usr := range serv.Users {
		usrName := usr
		dgUsr, err := s.GuildMember(serv.ID, usr)
		if err == nil {
			usrName = dgUsr.Nick
		}
		lb = append(lb, scoreboard{usr, usrName, b.GetUserScore(usr, serv), ""})
	}
	sort.Slice(lb, func(i, j int) bool {
		return lb[i].score > lb[j].score
	})
	if len(lb) > 0 {
		rank := 1
		lb[0].rank = rankString(rank)
		for i := 1; i < len(lb); i++ {
			prev := lb[i-1]
			cur := lb[i]
			if prev.score != cur.score {
				rank = i + 1
			}
			cur.rank = rankString(rank)
			lb[i] = cur
		}
	}
	return lb
}

func Leaderboard(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == b.UserID {
			return
		}
		serv := b.GetServer(m.GuildID)
		channel := m.ChannelID

		_, ok := b.triggered(s, m, serv.Prefix, cmd)
		if !ok {
			return
		}
		b.Info("command %s triggered", cmd)

		lb := b.getLeaderBoard(serv, s).Strings()
		menu := NewMenu(lb[1:], 10, channel, m.GuildID)
		menu.SetHeader(lb[0])
		menu.SetTitle("Server leaderboard")
		menu.SetSubtitle(fmt.Sprintf("Total number of points: `%d`", b.TotalPoints()))
		err := menu.Send(s)
		if err != nil {
			b.Error("creating menu: %s", err.Error())
			return
		}
		b.Menus[menu.ID()] = menu
		time.AfterFunc(time.Duration(61)*time.Second, purgeMenus(b, s))
	}
}

func (b *Bot) getRank(usr string, serv Server, s *DG.Session) string {
	lb := b.getLeaderBoard(serv, s)
	for _, sb := range lb {
		if sb.userID == usr {
			return sb.rank
		}
	}
	return rankString(len(lb) + 1)
}

func formatItemList(items []string) []string {
	lines := [][2]string{}
	maxCol1Width := 0
	for idx, item := range items {
		line := (idx/20)*10 + (idx % 10)
		column := (idx / 10) % 2
		if column == 0 {
			lines = append(lines, [2]string{item, ""})
			if len(item) > maxCol1Width {
				maxCol1Width = len(item)
			}
		} else {
			lines[line][1] = item
		}
	}

	res := []string{}
	col2Pos := maxCol1Width + 5
	for _, words := range lines {
		w1, w2 := words[0], words[1]
		res = append(res, w1+padLeft(w2, col2Pos-len(w1)))
	}
	return res
}

func Score(b *Bot, cmd string) func(*DG.Session, *DG.MessageCreate) {
	return func(s *DG.Session, m *DG.MessageCreate) {
		if m.Author.ID == b.UserID {
			return
		}
		serv := b.GetServer(m.GuildID)
		channel := m.ChannelID
		usr := m.Author.ID

		_, ok := b.triggered(s, m, serv.Prefix, cmd)
		if !ok {
			return
		}
		b.Info("command %s triggered", cmd)

		if _, ok := serv.Users[usr]; !ok {
			serv.Users[usr] = make([]string, 0)
			b.SaveServer(serv)
		}
		username := usr
		if dgUser, err := s.GuildMember(m.GuildID, usr); err == nil {
			username = dgUser.Nick
		}
		menu := NewMenu(formatItemList(serv.Users[usr]), 10, channel, m.GuildID)
		menu.SetTitle(username + "'s scoreboard")
		infos := fmt.Sprintf("Items: `%d/%d`", len(serv.Users[usr]), len(b.Items))
		infos += "⁠ ⁠ ⁠ ⁠ ⁠ " + fmt.Sprintf("Points: `%d`", b.GetUserScore(usr, serv))
		infos += "⁠ ⁠ ⁠ ⁠ ⁠ " + fmt.Sprintf("Rank: `%s`", b.getRank(usr, serv, s))
		menu.SetSubtitle(infos)
		err := menu.Send(s)
		if err != nil {
			b.Error("creating menu: %s", err.Error())
			return
		}
		b.Menus[menu.ID()] = menu
		time.AfterFunc(time.Duration(61)*time.Second, purgeMenus(b, s))
	}
}
