package main

import (
	"log"
	"net"
	"os"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://chat:chat@localhost:5432/chat"
	}
	store, err := newPostgresStore(dbURL)
	if err != nil {
		log.Fatalf("Unable to open database: %v", err)
	}

	s := newServer(store)
	go s.run()
	go runWebSocketBridge(":8080")

	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
	defer listener.Close()
	log.Printf("Server running on :8888")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Unable to accept connection: %v", err)
			continue
		}
		log.Printf("New client connected: %s", conn.RemoteAddr())

		c := &client{conn: conn, username: "Anonymous"}
		go c.readInput(s.commands)
	}
}
