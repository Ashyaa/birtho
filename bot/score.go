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
	user  string
	score int
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
	for pos, sb := range l {
		rank := rankString(pos + 1)
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

func getLeaderBoard(b *Bot, serv Server, s *DG.Session) []string {
	lb := leaderboard{}
	for usr := range serv.Users {
		usrName := usr
		dgUsr, err := s.GuildMember(serv.ID, usr)
		if err == nil {
			usrName = dgUsr.Nick
		}
		lb = append(lb, scoreboard{usrName, b.GetUserScore(usr, serv)})
	}
	sort.Slice(lb, func(i, j int) bool {
		return lb[i].score > lb[j].score
	})
	return lb.Strings()
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

		lb := getLeaderBoard(b, serv, s)
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
