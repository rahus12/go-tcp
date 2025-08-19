// let this be a simple tcp server
package main

import (
	"io"
	"log"
	"net"
	"sync"
)

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

	log.Println("waiting for reciever...")

	reciever, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("reciever connected from", reciever.RemoteAddr())

	log.Println("waiting for sender...")
	sender, err := ln.Accept()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("sender connected from", sender.RemoteAddr())

	var wg sync.WaitGroup
	wg.Add(2)

	//forward everything for sender to reciever

	// below is a way to initiate just before if checking
	go pipe(&wg, reciever, sender)
	go pipe(&wg, sender, reciever)

	wg.Wait()
	log.Println("connection pair closed")
}
