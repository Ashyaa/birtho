package bot

import (
	"fmt"
	"sort"
	"time"

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

func (b *Bot) GetUserScoreboard(user string, serv Server) ScoreBoard {
	serv = b.updateScore(user, serv)
	b.SaveServer(serv)
	var res ScoreBoard
	for _, sb := range serv.Lb {
		if sb.UID == user {
			return sb
		}
	}
	return res
}

type ScoreBoard struct {
	UID   string
	Name  string
	Score int
	Rank  string
}

type Leaderboard []ScoreBoard

// Sort the leaderboard by score in decreasing order, and updates the rank.
func (lb Leaderboard) sort() {
	sort.Slice(lb, func(i, j int) bool {
		return lb[i].Score > lb[j].Score
	})
	if len(lb) > 0 {
		rank := 1
		lb[0].Rank = rankString(rank)
		for i := 1; i < len(lb); i++ {
			prev := lb[i-1]
			cur := lb[i]
			if prev.Score != cur.Score {
				rank = i + 1
			}
			cur.Rank = rankString(rank)
			lb[i] = cur
		}
	}
}

func (lb Leaderboard) Strings() []string {
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
		rank := sb.Rank
		rank = padLeft(rank, rankPlaces-len(rank))

		score := fmt.Sprintf("%d", sb.Score)
		score = padLeft(score, 11-len(score))

		line := fmt.Sprintf("%s%s     %s", rank, score, sb.Name)
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
		for i := range serv.Lb {
			if serv.Lb[i].UID == member.User.ID {
				serv.Lb[i].Name = U.MemberName(member)
				break
			}
		}
	}
	return serv
}

func (b *Bot) updateScore(uid string, serv Server) Server {
	score := b.GetUserScore(uid, serv)
	found := false
	for i := range serv.Lb {
		if serv.Lb[i].UID != uid {
			continue
		}
		found = true
		if serv.Lb[i].Score == score {
			return serv
		}
		serv.Lb[i].Score = score
		break
	}
	if !found {
		serv.Lb = append(serv.Lb, ScoreBoard{UID: uid, Score: score})
		serv = b.updateLBNames(serv)
		return serv
	}
	serv.Lb.sort()
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

func (b *Bot) getLeaderBoard(serv Server) Leaderboard {
	if len(serv.Lb) == 0 || len(serv.Users) != len(serv.Lb) {
		lb := Leaderboard{}
		users := serv.Users
		for usr := range users {
			lb = append(lb, ScoreBoard{usr, "", b.GetUserScore(usr, serv), ""})
		}
		lb.sort()
		serv.Lb = lb
	}
	serv = b.updateLBNames(serv)
	b.SaveServer(serv)
	return serv.Lb
}

func ShowLeaderboard(b *Bot, p CommandParameters) {
	lb := b.getLeaderBoard(p.S).Strings()
	menu := NewMenu(lb[1:], 10, p.CID, p.GID)
	menu.SetHeader(lb[0])
	menu.SetTitle("Server leaderboard")
	subtitle := fmt.Sprintf("Total number of points: `%d`", b.TotalPoints())
	if p.S.G.Finished {
		subtitle += "\u2060 \u2060 \u2060 \u2060 \u2060 Winner: " + U.BuildUserTag(p.S.G.Winner)
	}
	menu.SetSubtitle(subtitle)
	err := menu.Send(b.s, p.I)
	if err != nil {
		b.Error("creating menu: %s", err.Error())
		return
	}
	b.Menus[menu.ID()] = menu
	time.AfterFunc(time.Duration(61)*time.Second, purgeMenus(b))
}

func (b *Bot) getItemList(usr string, serv Server) []string {
	res := []string{}
	userItems, ok := serv.Users[usr]
	if !ok {
		return res
	}
	return userItems
}

func formatItemList(monsters []Monster, playerItems []string) (res []string) {
	has := U.ToHashMap(playerItems)
	for _, monster := range monsters {
		res = append(res, monster.Name, "")
		for _, item := range monster.Items {
			_, found := has[item.ID]
			res = append(res, item.Description(!found))
		}
	}
	return res
}

const hint = "🔸Common\u2060 \u2060 \u2060 \u2060 \u2060 🟠Uncommon\u2060 \u2060 \u2060 \u2060 \u2060 🟧Rare"

func ShowScore(b *Bot, p CommandParameters) {
	sb := b.GetUserScoreboard(p.UID, p.S)
	monsters := b.SortedMonsters()
	images := []string{}
	for _, m := range monsters {
		images = append(images, m.URL)
	}
	menu := NewMenu(
		formatItemList(monsters, b.getItemList(p.UID, p.S)),
		5, p.CID, p.GID)
	menu.SetTitle(sb.Name + "'s scoreboard")
	menu.SetImages(images)
	infos := fmt.Sprintf("Items: `%d/%d`", len(p.S.Users[p.UID]), len(b.Items))
	infos += "\u2060 \u2060 \u2060 \u2060 \u2060 " + fmt.Sprintf("Points: `%d`", sb.Score)
	infos += "\u2060 \u2060 \u2060 \u2060 \u2060 " + fmt.Sprintf("Rank: `%s`", sb.Rank)
	menu.SetSubtitle(infos)
	menu.SetFooter(hint)
	err := menu.Send(b.s, p.I)
	if err != nil {
		b.Error("creating menu: %s", err.Error())
		return
	}
	b.Menus[menu.ID()] = menu
	time.AfterFunc(61*time.Second, purgeMenus(b))
}
