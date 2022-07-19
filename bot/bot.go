package bot

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	U "github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
	LR "github.com/sirupsen/logrus"
)

func New(log LR.Logger) (*Bot, error) {
	conf, err := ReadConfig(log)
	if err != nil {
		log.Error("error reading config: ", err)
		return nil, err
	}

	res := Bot{Log: log, Items: make(map[string]Item)}
	res.buildGameData(conf)

	res.dg, err = DG.New("Bot " + conf.Token)
	if err != nil {
		log.Error("error creating session: ", err)
		return nil, err
	}

	// Open the database
	res.OpenDB()

	// Install command handlers
	res.SetupCommands()

	// Provide a seed for the pseudo-random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	res.dg.Identify.Intents = DG.IntentsGuildMessages | DG.IntentGuildMessageReactions

	// Open a websocket connection to Discord and begin listening.
	err = res.dg.Open()
	if err != nil {
		res.ErrorE(err, "error opening connection")
		return nil, err
	}
	res.UserID = res.dg.State.User.ID
	res.Menus = make(map[string]Menu)
	res.Mention = U.BuildUserTag(res.UserID)

	return &res, nil
}

func (b *Bot) SetupCommands() {
	for _, cmd := range Commands {
		b.dg.AddHandler(cmd.Constructor(b, cmd.Name))
	}
	b.dg.AddHandler(PageReact(b))
}

func (b *Bot) Stop() {
	b.Info("closing database")
	err := b.db.Close()
	if err != nil {
		b.ErrorE(err, "closing database")
	}
}

// Returns true and the command content if the message triggers the command, else false and an empty string
func (b *Bot) triggered(s *DG.Session, m *DG.MessageCreate, prefix, command string) ([]string, bool) {
	tagCommand := b.Mention + " " + command
	payload := []string{}
	fields := strings.Fields(m.Content)
	if strings.HasPrefix(m.Content, tagCommand) {
		if len(m.Content) > len(tagCommand) && !strings.HasPrefix(m.Content, tagCommand+" ") {
			return payload, false
		}
		if len(fields) > 2 {
			payload = fields[2:]
		}
		return payload, true
	}
	prefixCommand := prefix + command
	if strings.HasPrefix(m.Content, prefixCommand) {
		if len(m.Content) > len(prefixCommand) && !strings.HasPrefix(m.Content, prefixCommand+" ") {
			return payload, false
		}
		if len(fields) > 1 {
			payload = fields[1:]
		}
		return payload, true
	}
	return payload, false
}

func (b *Bot) buildGameData(conf Config) {
	b.MonsterIds = make([]string, 0)
	b.Monsters = make(map[string]Monster)
	sum := 1
	for _, monster := range conf.Monsters {
		if monster.URL == "" {
			continue
		}
		key := strconv.Itoa(monster.ID)
		b.MonsterIds = append(b.MonsterIds, key)
		chance := int(monster.Chance * 100)
		monster.Range.min = sum
		monster.Range.max = sum + chance - 1
		if !monster.buildItems(b.Log) {
			b.Error("monster '%s' has no items and will be skipped", monster.Name)
			continue
		}
		b.AddItems(monster.Items)
		b.Monsters[key] = monster
		sum += chance
	}
	if sum-1 != 10000 {
		if sum > 1 {
			b.Warn("the sum of monster spawn chances is not equal to 100")
		}
		b.EqualMonsterChances = true
		b.Info("all monsters set to have equal chances to spawn")
	}
	if len(b.Monsters) == 0 {
		b.Fatal("no valid monsters in the configuration")
	}
}

func (b *Bot) AddItems(items []Item) {
	for _, item := range items {
		b.Items[item.Name] = item
	}
}

func (b *Bot) TotalPoints() int {
	res := 0
	for _, item := range b.Items {
		res += item.Points
	}
	return res
}
