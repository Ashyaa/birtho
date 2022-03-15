package bot

const DefaultPrefix = "b!"

type User struct {
	ID     string
	Caught []int
}

type Server struct {
	ID       string
	Prefix   string
	Channels []string
	Users    map[string]int
}
