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

	log.Println("connected as reciever; waiting for data....")

	if _, err := io.Copy(os.Stdout, conn); err != nil {
		log.Fatal(err)
	}
}
