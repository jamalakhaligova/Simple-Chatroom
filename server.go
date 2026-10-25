package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type server struct {
	rooms 		map[string]*room
	commands 	chan command
}

func newServer() *server {
	return &server{
		rooms: make(map[string]*room),
		commands: make(chan command),
	}
}

func (s *server) run() {
	for cmd := range s.commands {
		switch cmd.id {
		case CMD_USER:
			s.greet_user(cmd.client, cmd.args)
		case CMD_JOIN:
			s.join(cmd.client, cmd.args)
		case CMD_ROOMS:
			s.listRooms(cmd.client, cmd.args)
		case CMD_MSG:
			s.msg(cmd.client, cmd.args)
		case CMD_QUIT:
			s.quit(cmd.client)
		}
	}
}

func (s *server) newClient(conn net.Conn) *client {
	log.Printf("New client connected: %s", conn.RemoteAddr().String())
	
	return &client{
		conn: conn,
		username: "Anonymous",
		commands: s.commands,
	}
}

func (s *server) greet_user(c *client, args []string){
	if len(args) < 2 {
		c.msg("Username is required")
		return
	}
	c.username = args[1]
	c.msg(fmt.Sprintf("Hi %s", c.username))
}

func (s *server) join(c *client, args []string){
	if len(args) < 2 {
		c.msg("Room name is required. Use: /join room")
		return
	}
	roomName := args[1]
	r, ok := s.rooms[roomName]
	if !ok {
		r = &room{
			name: roomName,
			members: make(map[net.Addr]*client),
		}
		s.rooms[roomName] = r
	}

	r.members[c.conn.RemoteAddr()] = c
	s.quitCurrentRoom(c)
	c.room = r
	r.broadcast(c, fmt.Sprintf("%s joined the room", c.username))
	c.msg(fmt.Sprintf("Welcome to %s", roomName))
	
}

func (s *server) listRooms(c *client, args []string){
	var rooms []string
	for name := range s.rooms {
		rooms = append(rooms, name)
	}
	c.msg(fmt.Sprintf("Available rooms are: %s", strings.Join(rooms, ", ")))
}

func (s *server) msg(c *client, args []string){
	if len(args) < 2 {
		c.msg("Message is required")
		return
	}
	if c.room == nil {
		c.msg("You must join the room first")
		return
	}

	c.room.broadcast(c, c.username + ": " + strings.Join(args[1:len(args)], " "))
}

func (s *server) quit(c *client){
	log.Printf("Client disconnected: %s", c.conn.RemoteAddr().String())

	s.quitCurrentRoom(c)
	c.msg(fmt.Sprintf("Goodbye %s", c.username))
	c.conn.Close()
}

func (s *server) quitCurrentRoom(c *client){
	if c.room != nil {
		delete(c.room.members, c.conn.RemoteAddr())
		c.room.broadcast(c, fmt.Sprintf("%s left the chatroom", c.username))
	}
}

