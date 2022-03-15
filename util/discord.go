package util

import (
	"fmt"
	"regexp"

	DG "github.com/bwmarrin/discordgo"
)

var ChannelTagPattern = regexp.MustCompile("<#([0-9]{18})>")

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
