package main

// assigning command id to mean int
type commandID int

// declaring constants which will be out commands
const (
	CMD_NICK commandID = iota // iota basically makes this 0 and all other increment by 1, so 0,1,2,3,...
	CMD_JOIN
	CMD_ROOMS
	CMD_MSG
	CMD_QUIT
)

type command struct {
	id     commandID
	client *client
	args   []string
}
