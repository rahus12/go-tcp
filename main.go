// must first run go build

package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	s := newServer()

	// start a goruntine to run the server
	go s.run()

	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server started, listening on port 8888")
	defer listener.Close()

	// same inf loop to constantly accept new connections
	for {
		conn, err := listener.Accept() //remember this is a blocking call
		if err != nil {
			log.Println("Error accepting clients...", err)
			continue
		}

		// if all is well create a new client
		c := s.newClient(conn)

		// read the inputs
		go c.readInput()

	}
}
