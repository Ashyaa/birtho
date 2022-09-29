package bot

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

func (b *Bot) GetUserScore(user string, serv Server) int {
	res := 0
	if _, ok := serv.Users[user]; !ok {
		serv.Users[user] = make([]string, 0)
	}
	for _, itemID := range serv.Users[user] {
		if item, ok := b.Items[itemID]; ok {
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
	res := "\u2060" + s

	for i := 0; i < padding; i++ {
		res = " " + res
	}
	return res
}

func (b *Bot) getLeaderBoard(serv Server) leaderboard {
	lb := leaderboard{}
	for usr := range serv.Users {
		usrName := usr
		dgUsr, err := b.s.GuildMember(serv.ID, usr)
		if err == nil {
			if dgUsr.Nick != "" {
				usrName = dgUsr.Nick
			} else {
				usrName = dgUsr.User.Username
			}
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

func Leaderboard(b *Bot, p CommandParameters) {
	lb := b.getLeaderBoard(p.S).Strings()
	menu := NewMenu(lb[1:], 10, p.CID, p.GID)
	menu.SetHeader(lb[0])
	menu.SetTitle("Server leaderboard")
	menu.SetSubtitle(fmt.Sprintf("Total number of points: `%d`", b.TotalPoints()))
	err := menu.Send(b.s, p.I)
	if err != nil {
		b.Error("creating menu: %s", err.Error())
		return
	}
	b.Menus[menu.ID()] = menu
	time.AfterFunc(time.Duration(61)*time.Second, purgeMenus(b))
}

func (b *Bot) getRank(usr string, serv Server) string {
	lb := b.getLeaderBoard(serv)
	for _, sb := range lb {
		if sb.userID == usr {
			return sb.rank
		}
	}
	return rankString(len(lb) + 1)
}

func wrap(in string, width int) (r1 string, r2 string) {
	strs := strings.Split(in, " ")
	r1, r2 = strs[0], ""
	len := utf8.RuneCountInString
	curLen := len(r1)
	for _, s := range strs[1:] {
		runeCount := len(s)
		if curLen+runeCount+1 <= width {
			r1 += " " + s
			curLen += runeCount + 1
		} else {
			if len(r2) > 0 {
				r2 += " "
			}
			r2 += s
		}
	}
	return
}

func (b *Bot) getItemList(usr string, serv Server) []string {
	res := []string{}
	for _, itemID := range serv.Users[usr] {
		if item, ok := b.Items[itemID]; ok {
			res = append(res, item.Name)
		}
	}
	return res
}

func formatItemList(items []string) []string {
	lines := [][2]string{}
	len := utf8.RuneCountInString
	for idx, item := range items {
		line := (idx/20)*10 + (idx % 10)
		column := (idx / 10) % 2
		if column == 0 {
			lines = append(lines, [2]string{item, ""})
		} else {
			lines[line][1] = item
		}
	}

	res := []string{}
	maxCol1Width := 26
	col2Pos := maxCol1Width + 4
	for _, words := range lines {
		w1_1, w1_2 := wrap(words[0], maxCol1Width)
		w2_1, w2_2 := wrap(words[1], maxCol1Width)
		line := w1_1 + padLeft(w2_1, col2Pos-len(w1_1))
		if len(w1_2) > 0 || len(w2_2) > 0 {
			line += "\n" + w1_2 + padLeft(w2_2, col2Pos-len(w1_2))
		}
		res = append(res, line, "  ")
	}
	return res
}

func Score(b *Bot, p CommandParameters) {
	if _, ok := p.S.Users[p.UID]; !ok {
		p.S.Users[p.UID] = make([]string, 0)
		b.SaveServer(p.S)
	}
	var username string
	if dgUser, err := b.s.GuildMember(p.GID, p.UID); err == nil {
		if dgUser.Nick != "" {
			username = dgUser.Nick
		} else {
			username = dgUser.User.Username
		}
	}
	itemList := b.getItemList(p.UID, p.S)
	menu := NewMenu(formatItemList(itemList), 20, p.CID, p.GID)
	menu.SetTitle(username + "'s scoreboard")
	infos := fmt.Sprintf("Items: `%d/%d`", len(p.S.Users[p.UID]), len(b.Items))
	infos += "\u2060 \u2060 \u2060 \u2060 \u2060 " + fmt.Sprintf("Points: `%d`", b.GetUserScore(p.UID, p.S))
	infos += "\u2060 \u2060 \u2060 \u2060 \u2060 " + fmt.Sprintf("Rank: `%s`", b.getRank(p.UID, p.S))
	menu.SetSubtitle(infos)
	err := menu.Send(b.s, p.I)
	if err != nil {
		b.Error("creating menu: %s", err.Error())
		return
	}
	b.Menus[menu.ID()] = menu
	time.AfterFunc(time.Duration(61)*time.Second, purgeMenus(b))
}
