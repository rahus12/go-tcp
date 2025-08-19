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
	log.Println("connected as sender; type and press Enter (Ctrl+D to end)")

	if _, err := io.Copy(conn, os.Stdin); err != nil {
		log.Fatal(err)
	}
}
