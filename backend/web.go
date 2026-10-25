package main

import (
	"io"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/websocket"
)

// Browsers can't open raw TCP connections, so each browser websocket
// gets its own TCP connection to the chat server and data is copied both ways.
func bridge(ws *websocket.Conn) {
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		log.Printf("Unable to reach chat server: %v", err)
		ws.Close()
		return
	}

	go func() {
		io.Copy(ws, conn)
		ws.Close()
	}()
	io.Copy(conn, ws)
	conn.Close()
}

func runWebSocketBridge(addr string) {
	http.Handle("/ws", websocket.Handler(bridge))

	log.Printf("WebSocket bridge running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
