package main

import (
	"bufio"
	"net"
	"strings"
)

type client struct {
	conn     net.Conn
	username string
	room     *room
}

// command is one line typed by a client, e.g. "/msg hello world"
// becomes {name: "/msg", arg: "hello world"}.
type command struct {
	client *client
	name   string
	arg    string
}

// readInput sends every line the client types to the server as a command.
func (c *client) readInput(commands chan<- command) {
	reader := bufio.NewReader(c.conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// connection dropped without /quit
			commands <- command{client: c, name: "/quit"}
			return
		}

		name, arg, _ := strings.Cut(strings.TrimSpace(line), " ")
		commands <- command{client: c, name: name, arg: strings.TrimSpace(arg)}
		if name == "/quit" {
			return
		}
	}
}

func (c *client) send(msg string) {
	c.conn.Write([]byte(">: " + msg + "\n"))
}

func (c *client) sendError(msg string) {
	c.conn.Write([]byte("Error: " + msg + "\n"))
}
