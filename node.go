package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

func notmain() {
	// now passing args which has names so check
	if len(os.Args) < 2 {
		log.Fatal("usage: node <username>")
	}
	username := os.Args[1]

	conn, err := net.Dial("tcp", "127.0.0.1:9000")

	if err != nil {
		log.Fatal(err)
	}
	// Send username as the first line (handshake)
	if _, err := fmt.Fprintln(conn, username); err != nil {
		log.Fatal("failed to send username:", err)
	}

	defer conn.Close()
	log.Println("Connected, Say Hi!")

	// we are kind of type casting conn to *net.TCPConn so we can use its methods even if we do not get it from net.dial
	tcp, _ := conn.(*net.TCPConn)

	// var wg sync.WaitGroup
	// wg.Add(2) // we dont need wait groups as we have decided to use channels instead

	//empty structs need 0 memory so this is a lightweight way to signal done
	done := make(chan struct{})
	var once sync.Once
	signal := func() { once.Do(func() { close(done) }) }

	//stdin -> conn
	go func() {
		// io.copy is a blocking code
		if _, err := io.Copy(conn, os.Stdin); err != nil {
			log.Println("write error: ", err)
		}
		// signal no more writes from this side
		if tcp != nil {
			_ = tcp.CloseWrite()
		}
		signal() // tell main that one direction finished -> chamnnel
	}()

	// conn -> stdout
	go func() {
		// defer wg.Done()
		if _, err := io.Copy(os.Stdout, conn); err != nil {
			log.Fatal(err)
		}
		// optional: stop further reads
		if tcp != nil {
			_ = tcp.CloseRead()
		}
		signal() // tell main that one direction finished
	}()

	// // this has to be below the above go function, as this will block the thread
	// now code changed and no need of below
	// if _, err := io.Copy(os.Stdout, conn); err != nil {
	// 	log.Fatal(err)
	// }

	//wg.Wait() // not needed since we are using channels

	// if we do not have this, then main function wont wait for go-routines to close and shuts off, whihc also closes other go routines thus
	// ending the program
	<-done
}
