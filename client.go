package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type client struct {
	conn     net.Conn //the tcp connection
	nick     string
	room     *room
	commands chan<- command // declares a send only client

}

// this will be a blocking function but
// doesnt matter as we have called a goroutine in the main

// this part was moved out form the oldmain to its own seperate function
func (c *client) readInput() {
	// inf loop to read messaegs
	for {
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			return
		}

		msg = strings.Trim(msg, "\r\n") // remove \n\r

		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0]) // remove spaces

		// we use this to send commands using the commands channel
		switch cmd {
		case "/nick":
			c.commands <- command{
				id:     CMD_NICK,
				client: c,
				args:   args,
			}
		case "/join":
			c.commands <- command{
				id:     CMD_JOIN,
				client: c,
				args:   args,
			}
		case "/rooms":
			c.commands <- command{
				id:     CMD_ROOMS,
				client: c,
				args:   args,
			}
		case "/msg":
			c.commands <- command{
				id:     CMD_MSG,
				client: c,
				args:   args,
			}
		case "/quit":
			c.commands <- command{
				id:     CMD_QUIT,
				client: c,
				args:   args,
			}
		default:
			c.err(fmt.Errorf("unkown command %s", cmd)) // need to write the err
		}
	}
}

func (c *client) err(err error) {
	c.conn.Write([]byte("ERR: " + err.Error() + "\n"))
}

func (c *client) msg(msg string) {
	c.conn.Write(([]byte("> " + msg + "\n")))
}
