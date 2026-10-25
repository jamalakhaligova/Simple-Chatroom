package main

type room struct {
	name    string
	members map[*client]bool
}

func newRoom(name string) *room {
	return &room{name: name, members: make(map[*client]bool)}
}

// broadcast sends msg to everyone in the room except the sender.
func (r *room) broadcast(sender *client, msg string) {
	for member := range r.members {
		if member != sender {
			member.send(msg)
		}
	}
}
