package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:9000")

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	log.Println("Connected, Say Hi!")

	go func() {
		if _, err := io.Copy(conn, os.Stdin); err != nil {
			log.Println("write error: ", err)
		}
	}()

	// this has to be below the above go function, as this will block the thread
	if _, err := io.Copy(os.Stdout, conn); err != nil {
		log.Fatal(err)
	}
}
