package bot

import "github.com/asdine/storm/v3"

func (b *Bot) OpenDB() {
	var err error
	b.db, err = storm.Open("app.db")
	if err != nil {
		b.FatalE(err, "opening database")
	}
}

func (b *Bot) GetServer(id string) Server {
	var res Server
	err := b.db.One("ID", id, &res)
	if err != nil {
		return b.NewServer(id)
	}
	// Safety checks
	if res.G.Monsters == nil {
		res.G.Monsters = make(map[string]MonsterSpawn)
	}
	if res.Users == nil {
		res.Users = make(map[string][]string)
	}
	if IsHistoryInvalid(res.G.LastMessages) {
		res.G.LastMessages = NewHistory()
	}
	return res
}

func (b *Bot) NewServer(id string) Server {
	serv := Server{
		ID:     id,
		Prefix: DefaultPrefix,
		G: Game{
			Monsters:      make(map[string]MonsterSpawn),
			MinDelay:      DefaultMinDelay,
			VariableDelay: DefaultVariableDelay,
			StayTime:      DefaultStayTime,
			LastMessages:  NewHistory(),
		},
		Channels: make([]string, 0),
		Admins:   make([]string, 0),
		Users:    make(map[string][]string),
		Lb:       make(Leaderboard, 0),
	}
	b.db.Save(&serv)
	return serv
}

func (b *Bot) SaveServer(s Server) {
	b.db.Save(&s)
}
