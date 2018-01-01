package message

const (
	Join = iota
	Shake
)

type Msg struct {
	Username string `json:"username"`
	Code     int    `json:"code"`
}

func New(username string, code int) *Msg {
	return &Msg{Username: username, Code: code}
}
