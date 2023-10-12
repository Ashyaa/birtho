package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	U "github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
)

func (b *Bot) RandomMonster() Monster {
	if b.EqualMonsterChances {
		index := b.rng.Intn(len(b.MonsterIds))
		key := b.MonsterIds[index]
		return b.Monsters[key]
	}
	number := b.rng.Intn(10000) + 1
	for _, m := range b.Monsters {
		if m.Range.Belongs(number) {
			return m
		}
	}
	b.Fatal("invalid monster roll")
	return Monster{}
}

func Play(b *Bot, p CommandParameters) {
	arg := strings.ToLower(p.Options["state"].(string))
	if arg != "on" && arg != "off" {
		msg := fmt.Sprintf("usage: `%s%s <on|off>`", p.S.Prefix, p.Name)
		SendText(b.s, p.I, p.CID, msg)
		return
	}

	p.S.Cooldown(b.rng)
	p.S.G.On = arg == "on"
	b.SaveServer(p.S)

	status := "off"
	if p.S.G.On {
		status = "on"
	}

	msg := fmt.Sprintf("Play status set to `%s`", status)
	SendText(b.s, p.I, p.CID, msg)
}

func Reset(b *Bot, p CommandParameters) {
	p.S.Users = make(map[string][]string)
	p.S.G.Finished = false
	p.S.G.Winner = ""
	b.SaveServer(p.S)

	msg := "Cleared all players' item list and reset the game status!"
	SendText(b.s, p.I, p.CID, msg)
}

func Spawn(b *Bot, p CommandParameters) {
	isManualCommand := p.S.IsAdmin(p.UID) && p.IsUserTriggered
	if !p.S.CanSpawn(p.CID) && !isManualCommand {
		return
	}
	if isManualCommand {
		b.Info("command %s triggered manually", p.Name)
	} else {
		if b.TriggersAnyOtherCommand(p) {
			return
		}
		if !p.S.G.Spawns(b.rng) {
			p.S.G.IncreaseSpawnRate(b.rng, p.MsgCreate.Message)
			b.SaveServer(p.S)
			return
		}
		userName := p.UID
		member, err := b.s.GuildMember(p.GID, p.UID)
		if err == nil {
			userName = U.MemberName(member)
		}
		b.Info("command %s triggered by %s", p.Name, userName)
		p.S.G.LastMessages = p.S.G.LastMessages.Update(p.MsgCreate.Message)
	}

	p.S.G.ResetSpawnRate()

	monster := b.RandomMonster()
	spawn := MonsterSpawn{
		ID:       strconv.Itoa(monster.ID),
		Expected: "trick",
	}

	if trickOrTreat(b.rng) {
		spawn.Expected = "treat"
	}
	command := p.S.Prefix + spawn.Expected
	footer := fmt.Sprintf("Post \"%s\" to get an item!\nArt by %s.", command, monster.Artist)

	msg, err := SendEmbed(b.s, p.I, p.CID, embed.NewEmbed().
		SetTitle("A visitor has come!").
		SetDescription(fmt.Sprintf("**%s** appeared! Greet them with `%s`!", monster.Name, command)).
		SetColor(0x00FF00).
		SetFooter(footer).
		SetImage(monster.URL).MessageEmbed,
		nil,
	)
	if err != nil {
		b.ErrorE(err, "spawn message")
		return
	}
	spawn.Message = msg.ID
	p.S.G.Monsters[p.CID] = spawn
	p.S.Cooldown(b.rng)
	b.SaveServer(p.S)

	time.AfterFunc(p.S.G.StayTime, func() {
		curServ := b.GetServer(p.GID)
		spawn, ok := curServ.G.Monsters[p.CID]
		b.Info("spawn message: %s", spawn.Message)
		b.Info("actual message: %s", msg.ID)
		if !ok || spawn.Message != msg.ID {
			return
		}
		delete(curServ.G.Monsters, p.CID)
		b.SaveServer(curServ)

		edit := DG.NewMessageEdit(p.CID, msg.ID).SetEmbed(embed.NewEmbed().
			SetTitle("The visitor has left.").
			SetDescription(fmt.Sprintf("**%s** left...", monster.Name)).
			SetColor(0xFF0000).MessageEmbed)
		b.s.ChannelMessageEditComplex(edit)
	})
}

const winningMessage = "Congratulations %s! You gathered all items and became the one true Spooky Lord!"

func Grab(b *Bot, p CommandParameters) {
	channel := p.CID
	if !U.Contains(p.S.Channels, channel) {
		return
	}

	spawn, ok := p.S.G.Monsters[channel]
	if !ok {
		return
	}

	if _, ok = p.S.Users[p.UID]; !ok {
		p.S.Users[p.UID] = make([]string, 0)
	}

	if p.Name == spawn.Expected {
		monster := b.Monsters[spawn.ID]
		item := monster.RandomItem(b.rng, b.Log)
		text := fmt.Sprintf("As a thank you for your kindness, **%s** gives %s one **%s**",
			monster.Name, U.BuildUserTag(p.UID), item.Description(false))
		duplicate := U.Contains(p.S.Users[p.UID], item.ID)
		footer := itemDescription(item, duplicate) + "\n" + fmt.Sprintf("Art by %s.", monster.Artist)
		b.s.ChannelMessageEditEmbed(channel, spawn.Message, embed.NewEmbed().
			SetTitle("The visitor has been pleased!").
			SetDescription(text).
			SetColor(0xFFFFFF).
			SetFooter(footer).
			SetImage(monster.URL).MessageEmbed)
		b.s.MessageReactionAdd(p.CID, p.MsgCreate.ID, "✅")
		p.S.Users[p.UID] = U.AppendUnique(p.S.Users[p.UID], item.ID)
		if !duplicate {
			p.S = b.updateScore(p.UID, p.S)
		}
		if !p.S.G.Finished && b.GetUserScore(p.UID, p.S) == b.TotalPoints() {
			p.S.G.Finished = true
			p.S.G.Winner = p.UID
			SendText(b.s, p.I, p.CID, fmt.Sprintf(winningMessage, U.BuildUserTag(p.UID)))
		}
	} else {
		monster := b.Monsters[spawn.ID]
		text := fmt.Sprintf("%s scared **%s** away...", U.BuildUserTag(p.UID), monster.Name)
		b.s.ChannelMessageEditEmbed(channel, spawn.Message, embed.NewEmbed().
			SetTitle("The visitor has fled!").
			SetDescription(text).
			SetColor(0xFF0000).MessageEmbed)
		b.s.MessageReactionAdd(p.CID, p.MsgCreate.ID, "❌")
	}
	delete(p.S.G.Monsters, channel)
	b.SaveServer(p.S)
}

func itemDescription(item Item, duplicate bool) string {
	text := ""
	if item.Chance < 20 {
		text = "🟧This item is rare. It must be worth a lot."
	} else if item.Chance < 50 {
		text = "🟠This item is uncommon. You wonder where they got it..."
	} else {
		text = "🔸This item is common. There's nothing special about it."
	}
	if duplicate {
		return text + " Sadly, you already had one..."
	}
	return text + " It has been added to your inventory."
}

func trickOrTreat(rng U.RNG) bool {
	n := rng.Intn(1000)
	return n > 499
}

func GiveNew(b *Bot, p CommandParameters) {
	items := []string{}
	for item := range b.Items {
		items = append(items, item)
	}

	if _, ok := p.S.Users[p.UID]; !ok {
		p.S.Users[p.UID] = make([]string, 0)
	}
	if U.Contains(p.S.Users[p.UID], "") {
		newItems := []string{}
		for _, item := range p.S.Users[p.UID] {
			if item == "" {
				continue
			}
			newItems = append(newItems, item)
		}
		p.S.Users[p.UID] = newItems
	}
	if len(p.S.Users[p.UID]) == len(items) {
		SendText(b.s, p.I, p.CID, "Already have all items")
		return
	}
	var item string
	for {
		item = items[b.rng.Intn(len(items))]
		if !U.Contains(p.S.Users[p.UID], item) {
			break
		}
	}
	p.S.Users[p.UID] = U.AppendUnique(p.S.Users[p.UID], item)
	msg := fmt.Sprintf("Gave you one `%s`", b.Items[item].Name)
	if len(p.S.Users[p.UID]) == len(items) {
		p.S.G.Finished = true
		p.S.G.Winner = p.UID
		SendText(b.s, p.I, p.CID, fmt.Sprintf(winningMessage, U.BuildUserTag(p.UID)))
	}
	b.SaveServer(p.S)

	SendText(b.s, p.I, p.CID, msg)
}

const (
	// Help texts
	description = "This bot is a game similar to the Halloween bot that Discord made available in Halloween 2020. Spooky stuff ahead!"
	howToPlay   = "When speaking in %s, spooky visitors may appear. Try to get their attention them by using the `%strick` or `%streat` commands! " +
		"Only the fastest person will get the chance to please them, and if pleased, they will rewards you with an item. " +
		"The first person to get all the items wiill be declared the winner!"
	visitorsAndItems = "Different kinds of visitors may appear, and each one of them may reward you with three different items. " +
		"Some items are more common than others. Rare items are worth more points."
	scoring = "See your score, rank in the leaderboard and current list of items with the `%sscore` command, or the `/score` slash command!\n" +
		"See the server leaderboard with the `%sleaderboard` command, or the `/leaderboard` slash command!"
	help   = "You can display this help message at any time wit the `%shelp` command, or the `/help` slash command."
	footer = "Developped by Ashyaa. Art by AcidFiend,  BirthofVns, Ella, N1MH and Talondal."
)

func Help(b *Bot, p CommandParameters) {
	// Create new embed message
	msg := embed.NewEmbed().SetTitle("Help").SetColor(0x008080)
	msg.AddField("What is this bot?", description)

	msg.AddField("How to play?", fmt.Sprintf(howToPlay, listChannels(p.S.Channels), p.S.Prefix, p.S.Prefix))

	msg.AddField("Visitors and items", visitorsAndItems)

	msg.AddField("Scoreboard and server leaderboard", fmt.Sprintf(scoring, p.S.Prefix, p.S.Prefix))

	msg.AddField("Help", fmt.Sprintf(help, p.S.Prefix))

	msg.SetFooter(footer)

	SendEmbed(b.s, p.I, p.CID, msg.MessageEmbed, nil)
}

func listChannels(channels []string) string {
	if len(channels) == 0 {
		return "any channel"
	}
	if len(channels) == 1 {
		return U.BuildChannelTag(channels[0])
	}
	last := channels[len(channels)-1]
	list := []string{}
	for _, c := range channels[:len(channels)-1] {
		list = append(list, U.BuildChannelTag(c))
	}
	return strings.Join(list, ", ") + " and " + U.BuildChannelTag(last)
}
