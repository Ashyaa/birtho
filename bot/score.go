package bot

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	U "github.com/ashyaa/birtho/util"
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

// Sort the leaderboard by score in decreasing order, and updates the rank.
func (lb leaderboard) sort() {
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
}

func (lb leaderboard) Strings() []string {
	res := []string{}
	rankPlaces := len(fmt.Sprintf("%d", len(lb))) + 2 // + suffix 'st', 'nd', 'rd', 'th'
	if rankPlaces < 4 {
		rankPlaces = 4
	}
	header := "Rank     Points     User"
	if rankPlaces > 4 {
		header = padLeft(header, rankPlaces-4)
	}
	res = append(res, header)
	if len(lb) == 0 {
		return res
	}
	for _, sb := range lb {
		rank := sb.rank
		rank = padLeft(rank, rankPlaces-len(rank))

		score := fmt.Sprintf("%d", sb.score)
		score = padLeft(score, 11-len(score))

		line := fmt.Sprintf("%s%s     %s", rank, score, sb.user)
		res = append(res, line)
	}
	return res
}

// Update leaderboard names with current user nicknames if any, else username
func (b *Bot) updateLBNames(serv Server) Server {
	members, err := b.s.GuildMembers(serv.ID, "", 1000)
	if err != nil {
		return serv
	}
	for _, member := range members {
		if _, ok := serv.Users[member.User.ID]; !ok {
			continue
		}
		usrName := member.User.Username
		if member.Nick != "" {
			usrName = member.Nick
		}
		for i := range serv.lb {
			if serv.lb[i].userID == member.User.ID {
				serv.lb[i].user = usrName
				break
			}
		}
	}
	return serv
}

func (b *Bot) updateScore(uid string, serv Server) Server {
	score := b.GetUserScore(uid, serv)
	found := false
	for i := range serv.lb {
		if serv.lb[i].userID != uid {
			continue
		}
		found = true
		serv.lb[i].score = score
	}
	if !found {
		serv.lb = append(serv.lb, scoreboard{userID: uid, score: score})
		serv = b.updateLBNames(serv)
	}
	serv.lb.sort()
	return serv
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
	if len(serv.lb) == 0 || len(serv.Users) != len(serv.lb) {
		lb := leaderboard{}
		users := serv.Users
		for usr := range users {
			lb = append(lb, scoreboard{usr, "", b.GetUserScore(usr, serv), ""})
		}
		lb.sort()
		serv.lb = lb
	}
	b.updateLBNames(serv)
	b.SaveServer(serv)

	return serv.lb
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
	for _, sb := range serv.lb {
		if sb.userID == usr {
			return sb.rank
		}
	}
	return rankString(len(serv.lb) + 2)
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

const UnknownItem = "???   "

func (b *Bot) getItemList(usr string, serv Server) []string {
	res := []string{}
	userItems, ok := serv.Users[usr]
	if !ok {
		return res
	}
	for _, item := range b.SortedItems() {
		hasItem := U.Contains(userItems, item.ID)
		if hasItem {
			res = append(res, item.Description())
		} else {
			res = append(res, UnknownItem)
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
