package bot

import (
	"github.com/ashyaa/birtho/util"
)

const DefaultPrefix = "b!"

type Server struct {
	ID       string
	Prefix   string
	Playing  bool
	Channels []string
	Admins   []string
	Users    map[string][]int
}

func (s Server) IsAdmin(uid string) bool {
	// Everyone is an admin when no admin are registered
	if len(s.Admins) == 0 {
		return true
	}
	return util.Contains(s.Admins, uid)
}
