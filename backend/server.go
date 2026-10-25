package main

import (
	"log"
	"strings"
)

// server owns all rooms. Commands are handled one at a time in run(),
// so rooms and clients never need locks.
type server struct {
	rooms    map[string]*room
	commands chan command
	store    *store
}

func newServer(store *store) *server {
	s := &server{
		rooms:    make(map[string]*room),
		commands: make(chan command),
		store:    store,
	}

	names, err := store.loadRooms()
	if err != nil {
		log.Printf("Unable to load rooms: %v", err)
	}
	for _, name := range names {
		s.rooms[name] = newRoom(name)
	}
	return s
}

func (s *server) run() {
	for cmd := range s.commands {
		c := cmd.client
		switch cmd.name {
		case "/user":
			s.setName(c, cmd.arg)
		case "/join":
			s.join(c, cmd.arg)
		case "/rooms":
			s.listRooms(c)
		case "/msg":
			s.message(c, cmd.arg)
		case "/quit":
			s.quit(c)
		default:
			c.sendError("Unknown command: " + cmd.name)
		}
	}
}

func (s *server) setName(c *client, name string) {
	if name == "" {
		c.send("Username is required")
		return
	}
	c.username = name
	c.send("Hi " + name)
}

func (s *server) join(c *client, name string) {
	if name == "" {
		c.send("Room name is required. Use: /join room")
		return
	}

	r, ok := s.rooms[name]
	if !ok {
		r = newRoom(name)
		s.rooms[name] = r
		if err := s.store.saveRoom(name); err != nil {
			log.Printf("Unable to save room: %v", err)
		}
	}

	s.leaveRoom(c)
	r.members[c] = true
	c.room = r
	r.broadcast(c, c.username+" joined the room")
	c.send("Welcome to " + name)

	history, err := s.store.lastMessages(name, 20)
	if err != nil {
		log.Printf("Unable to load history: %v", err)
	}
	for _, msg := range history {
		c.send(msg)
	}
}

func (s *server) listRooms(c *client) {
	var names []string
	for name := range s.rooms {
		names = append(names, name)
	}
	c.send("Available rooms are: " + strings.Join(names, ", "))
}

func (s *server) message(c *client, text string) {
	if text == "" {
		c.send("Message is required")
		return
	}
	if c.room == nil {
		c.send("You must join the room first")
		return
	}

	if err := s.store.saveMessage(c.room.name, c.username, text); err != nil {
		log.Printf("Unable to save message: %v", err)
	}
	c.room.broadcast(c, c.username+": "+text)
}

func (s *server) quit(c *client) {
	log.Printf("Client disconnected: %s", c.conn.RemoteAddr())
	s.leaveRoom(c)
	c.send("Goodbye " + c.username)
	c.conn.Close()
}

func (s *server) leaveRoom(c *client) {
	if c.room == nil {
		return
	}
	delete(c.room.members, c)
	c.room.broadcast(c, c.username+" left the chatroom")
	c.room = nil
}
