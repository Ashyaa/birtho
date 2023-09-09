package bot

import (
	"fmt"
	"time"

	U "github.com/ashyaa/birtho/util"
	DG "github.com/bwmarrin/discordgo"
)

type Message struct {
	Time   time.Time
	Author string
	Valid  bool
}

func NewMessage(author string, time time.Time) Message {
	return Message{
		Author: author,
		Time:   time,
		Valid:  true,
	}
}

type History []Message

func NewHistory() History {
	res := make(History, HistoryDepth)
	now := time.Now()
	for i := 0; i < HistoryDepth; i++ {
		res[i] = NewMessage("", now)
	}
	return res
}

func IsHistoryInvalid(h History) bool {
	if len(h) != HistoryDepth {
		return true
	}
	for _, msg := range h {
		if !msg.Valid {
			return true
		}
	}
	return false
}

func (h History) Update(msg *DG.Message) History {
	msgTime, err := U.CreationTime(msg.ID)
	if err != nil {
		return h
	}
	res := append(h, NewMessage(msg.Author.ID, msgTime))
	return res[1:]
}

func (h History) Span() time.Duration {
	return h[HistoryDepth-1].Time.Sub(h[0].Time)
}

func (h History) HasSeveralAuthors() bool {
	auth := h[0].Author
	for i := 1; i < HistoryDepth; i++ {
		if h[i].Author != auth {
			return true
		}
	}
	return false
}

func (h History) String() string {
	res := "\n"
	for i, msg := range h {
		res += fmt.Sprintf("(%02d) %s - %s - %v\n", 10-i, msg.Author, msg.Time, msg.Valid)
	}
	return res
}
