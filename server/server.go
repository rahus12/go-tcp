// let this be a simple tcp server
package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

// to add a username to the clients, which also allows for rejoin
// fancy way of declaring multple variable instead of writing var again and again
var (
	clients = make(map[string]net.Conn) // username -> conn
	mu      sync.Mutex                  // protects clients
)

// read the first line as the username (trim newline)
func readUsername(c net.Conn) (string, error) {
	r := bufio.NewReader(c)
	name, err := r.ReadString('\n') // read till encounters first occ of delimiter
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(name), nil // like python strip, removes all whitespaces from both sides
}

func register(username string, c net.Conn) {
	mu.Lock()
	defer mu.Unlock() // same as writing it at the end but is prefered incase the program panics
	//ok is to check if the key exists or not
	if old, ok := clients[username]; ok && old != nil {
		_ = old.Close() // drop prev session
	}
	clients[username] = c
	log.Printf("user %q registered from %v\n", username, c.RemoteAddr())
}

// creating a single function named pipe to transfer data from one end ot another
// avoids creating seperate sender and reciever
// this now allows us to us goroutines and form bi-directional connections
// short hand for repeating types

func pipe(wg *sync.WaitGroup, dst, src net.Conn) {
	defer wg.Done()
	if _, err := io.Copy(dst, src); err != nil {
		log.Println("copy error: ", err)
	}
}

func main() {
	// need a listner
	ln, err := net.Listen("tcp", ":9000")

	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()
	log.Println("listening on port :9000")

	for {
		// Accept first node
		log.Println("waiting for first node...")
		//note: ln.ACcept is a blocking code
		aConn, err := ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		aName, err := readUsername(aConn)
		if err != nil || aName == "" {
			log.Println("failed to read username for first node:", err)
			_ = aConn.Close()
			continue
		}
		register(aName, aConn)

		// Accept second node
		log.Println("waiting for second node...")
		bConn, err := ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			_ = aConn.Close()
			continue
		}
		bName, err := readUsername(bConn)
		if err != nil || bName == "" {
			log.Println("failed to read username for second node:", err)
			_ = bConn.Close()
			_ = aConn.Close()
			continue
		}
		register(bName, bConn)

		log.Printf("pairing %q <-> %q\n", aName, bName)

		// Bidirectional piping for this pair
		var wg sync.WaitGroup
		wg.Add(2)
		go pipe(&wg, clients[bName], clients[aName]) // a -> b
		go pipe(&wg, clients[aName], clients[bName]) // b -> a

		// Wait until both directions are done
		wg.Wait()

		// Cleanup this pair (remove dead conns from map)
		mu.Lock()
		if clients[aName] != nil {
			_ = clients[aName].Close()
			clients[aName] = nil
		}
		if clients[bName] != nil {
			_ = clients[bName].Close()
			clients[bName] = nil
		}
		mu.Unlock()

		log.Printf("pair %q <-> %q closed; awaiting next pair...\n", aName, bName)
	}
}
