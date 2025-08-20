package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

type server struct {
	rooms    map[string]*room
	commands chan command
}

// this how u create a constructor to server
func newServer() *server {
	return &server{
		rooms:    make(map[string]*room),
		commands: make(chan command),
	}

}

// create a new client
func (s *server) newClient(conn net.Conn) *client {
	log.Printf("New Client has connected %s", conn.RemoteAddr().String())

	return &client{
		conn:     conn,
		nick:     "Anonymous",
		commands: s.commands,
	}

	// c.readInput()
}

// clients send messages which we need to read here on server side
func (s *server) run() {
	// channel opened for reading
	// loop over the channel and read
	for cmd := range s.commands {
		switch cmd.id {
		case CMD_NICK:
			s.nick(cmd.client, cmd.args) // defined below
		case CMD_JOIN:
			s.join(cmd.client, cmd.args) // need to create this
		case CMD_ROOMS:
			s.listRooms(cmd.client, cmd.args) // need to create this
		case CMD_MSG:
			s.msg(cmd.client, cmd.args) // need to create this
		case CMD_QUIT:
			s.quit(cmd.client, cmd.args) // need to create this
		}
	}
}

func (s *server) nick(c *client, args []string) {
	// simply set client.nick as nick
	c.nick = args[1] // 1 is nickname
	c.msg(fmt.Sprintf("Hello %s", c.nick))
}

func (s *server) join(c *client, args []string) {
	// need to check if the room exists
	// if not create the room
	// join to the room

	roomName := args[1]

	r, ok := s.rooms[roomName]
	if !ok {
		r = &room{
			name:    roomName,
			members: make(map[net.Addr]*client), // this is also net.addr ok both places
		}

		s.rooms[roomName] = r // map -> name : *room
	}

	r.members[c.conn.RemoteAddr()] = c // map -> remote Address : *client

	//quit the current room and then join the new
	s.quitCurrentRoom(c)
	c.room = r

	//announce new memeber arrival
	r.broadcast(c, fmt.Sprintf("%s has joined the room", c.nick))

	//welcome the new memeber
	c.msg(fmt.Sprintf("welcome to %s", r.name))

}

func (s *server) listRooms(c *client, args []string) {
	var rooms []string
	for name := range s.rooms {
		rooms = append(rooms, name)
	}

	c.msg(fmt.Sprintf("Available rooms: %s", strings.Join(rooms, ", ")))
}

// ask the client to join one room (any)
// but then i guess i must have a default room where everyone initally joins
func (s *server) msg(c *client, args []string) {
	if c.room == nil {
		c.err(errors.New("you must join a room first"))
		return
	}

	c.room.broadcast(c, c.nick+": "+strings.Join(args[1:], " "))
}

func (s *server) quit(c *client, args []string) {
	log.Printf("client has disconnected: %s", c.conn.RemoteAddr().String())

	s.quitCurrentRoom(c)

	//msg to client
	c.msg("sad to see you go")
	c.conn.Close()
}

func (s *server) quitCurrentRoom(c *client) {
	if c.room != nil {
		delete(c.room.members, c.conn.RemoteAddr()) // delete this address from the map called room.members
		// let all other memebers know this bitch got kicked or left
		c.room.broadcast(c, fmt.Sprintf("%s has left the room", c.nick))
	}
}
