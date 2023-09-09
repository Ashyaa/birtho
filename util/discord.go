package util

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	DG "github.com/bwmarrin/discordgo"
)

var ChannelTagPattern = regexp.MustCompile("<#([0-9]{18})>")
var UserTagPattern = regexp.MustCompile("<@!([0-9]{18})>")

// Return true and the channel ID if the input string matched the channel tag format
func StripChannelTag(cid string) (string, bool) {
	res := ChannelTagPattern.FindStringSubmatch(cid)
	if len(res) == 0 {
		return "", false
	}
	return res[1], true
}

// Returns the discord channel tag for channel with ID cid
func BuildChannelTag(cid string) string {
	return fmt.Sprintf("<#%s>", cid)
}

// Return true and the user ID if the input string matched the user tag format
func StripUserTag(uid string) (string, bool) {
	res := UserTagPattern.FindStringSubmatch(uid)
	if len(res) == 0 {
		return "", false
	}
	return res[1], true
}

// Returns the discord user tag for user with ID uid
func BuildUserTag(uid string) string {
	return fmt.Sprintf("<@!%s>", uid)
}

// Returns the discord user tag for user with ID uid
func Timestamp(t time.Time) string {
	return fmt.Sprintf("<t:%d:R>", t.Unix())
}

// Return true if cid is a valid channel in the guild identifed by gid
func IsValidChannel(s *DG.Session, gid, cid string) bool {
	channels, err := s.GuildChannels(gid)
	if err != nil {
		return false
	}
	for _, c := range channels {
		if c.ID == cid {
			return true
		}
	}
	return false
}

// Return true if the user uid is a member of the server gid
func IsUserInServer(s *DG.Session, gid, uid string) bool {
	_, err := s.GuildMember(gid, uid)
	return err == nil
}

func MemberName(m *DG.Member) string {
	if m == nil {
		return ""
	}
	if m.Nick != "" {
		return m.Nick
	}
	return m.User.Username
}

// CreationTime returns the creation time of a Snowflake ID relative to the creation of Discord.
// Taken from https://github.com/Moonlington/FloSelfbot/blob/master/commands/commandutils.go#L117
func CreationTime(ID string) (t time.Time, err error) {
	i, err := strconv.ParseInt(ID, 10, 64)
	if err != nil {
		return
	}
	timestamp := (i >> 22) + 1420070400000
	t = time.Unix(timestamp/1000, 0)
	return
}
