package main

import "net"

type room struct {
	name    string
	members map[net.Addr]*client
}

func (r *room) broadcast(sender *client, msg string) {
	for addr, member := range r.members {
		// avoids sending to the current client which left, but idk why its even needed as the client addr was already deleted
		if addr != sender.conn.RemoteAddr() {
			member.msg(msg)
		}
	}
}
