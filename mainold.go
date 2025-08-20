// let this be a simple tcp server
// evolved to named entity chat server
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
	clients    = make(map[string]net.Conn) // username -> conn
	mu         sync.Mutex                  // protects clients
	nameByConn = make(map[net.Conn]string) // conn -> username

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
	nameByConn[c] = username
	log.Printf("user %q registered from %v\n", username, c.RemoteAddr())
}

// creating a single function named pipe to transfer data from one end ot another
// avoids creating seperate sender and reciever
// this now allows us to us goroutines and form bi-directional connections
// short hand for repeating types
// evolve from pipe to pipelabelled which takes srcName to prepend it to show like alice: hi
// evolved from pipeLabelled to handleConn
func handleConn(wg *sync.WaitGroup, dst, src net.Conn, srcName string) {
	defer wg.Done()
	// if _, err := io.Copy(dst, src); err != nil {
	// 	log.Println("copy error: ", err)
	// }

	r := bufio.NewScanner(src)
	buf := make([]byte, 0, 64*1024) // 64KB initial size
	r.Buffer(buf, 1024*1024)        // max size 1MB

	for r.Scan() {
		line := strings.TrimRight(r.Text(), "\r\n")
		if line == "" {
			continue
		}

		_, err := io.WriteString(dst, srcName+": "+line+"\n")
		if err != nil {
			log.Println("write error:", err)
			return
		}
	}
	if err := r.Err(); err != nil && err != io.EOF {
		log.Println("read error:", err)
	}
}

func mainold() {

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
		// evolved from pipe to pipeLabeled
		go handleConn(&wg, clients[bName], clients[aName], aName) // a -> b
		go handleConn(&wg, clients[aName], clients[bName], bName) // b -> a
		// go handleConn(aConn)
		// go handleConn(bConn)

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
