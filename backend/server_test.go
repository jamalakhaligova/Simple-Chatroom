package main

import (
	"bytes"
	"net"
	"reflect"
	"strings"
	"testing"
)

// testDB keeps rooms and messages in memory instead of Postgres.
type testDB struct {
	rooms    []string
	messages map[string][]string
}

func newTestDB(rooms ...string) *testDB {
	return &testDB{rooms: rooms, messages: make(map[string][]string)}
}

func (db *testDB) saveRoom(name string) error {
	db.rooms = append(db.rooms, name)
	return nil
}

func (db *testDB) loadRooms() ([]string, error) {
	return db.rooms, nil
}

func (db *testDB) saveMessage(room, username, body string) error {
	db.messages[room] = append(db.messages[room], username+": "+body)
	return nil
}

func (db *testDB) lastMessages(room string, n int) ([]string, error) {
	return db.messages[room], nil
}

// testConn records what the server writes instead of sending it over the network.
type testConn struct {
	net.Conn
	out    bytes.Buffer
	closed bool
}

func (conn *testConn) Write(b []byte) (int, error) { return conn.out.Write(b) }
func (conn *testConn) Close() error                { conn.closed = true; return nil }
func (conn *testConn) RemoteAddr() net.Addr        { return &net.TCPAddr{} }

func newClient(name string) (*client, *testConn) {
	conn := &testConn{}
	return &client{conn: conn, username: name}, conn
}

// received returns the lines sent to a client since the last call.
func received(conn *testConn) []string {
	out := strings.TrimSuffix(conn.out.String(), "\n")
	conn.out.Reset()
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

func expect(t *testing.T, conn *testConn, want ...string) {
	t.Helper()
	if got := received(conn); !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestSetUsername(t *testing.T) {
	s := newServer(newTestDB())
	c, conn := newClient("Anonymous")

	s.handle(command{client: c, name: "/user"})
	expect(t, conn, ">: Username is required")

	s.handle(command{client: c, name: "/user", arg: "alice"})
	expect(t, conn, ">: Hi alice")
}

func TestJoinRoom(t *testing.T) {
	db := newTestDB()
	db.messages["go"] = []string{"carol: earlier message"}
	s := newServer(db)
	alice, aliceConn := newClient("alice")
	bob, bobConn := newClient("bob")

	s.handle(command{client: alice, name: "/join"})
	expect(t, aliceConn, ">: Room name is required. Use: /join room")

	s.handle(command{client: alice, name: "/join", arg: "go"})
	expect(t, aliceConn, ">: Welcome to go", ">: carol: earlier message")

	s.handle(command{client: bob, name: "/join", arg: "go"})
	expect(t, aliceConn, ">: bob joined the room")
	expect(t, bobConn, ">: Welcome to go", ">: carol: earlier message")

	if !reflect.DeepEqual(db.rooms, []string{"go"}) {
		t.Fatalf("expected room go to be saved once, got %v", db.rooms)
	}

	// switching rooms leaves the old one
	s.handle(command{client: bob, name: "/join", arg: "rust"})
	expect(t, aliceConn, ">: bob left the chatroom")
	expect(t, bobConn, ">: Welcome to rust")
}

func TestListRooms(t *testing.T) {
	s := newServer(newTestDB("saved"))
	c, conn := newClient("alice")

	s.handle(command{client: c, name: "/rooms"})
	expect(t, conn, ">: Available rooms are: saved")
}

func TestSendMessage(t *testing.T) {
	db := newTestDB()
	s := newServer(db)
	alice, aliceConn := newClient("alice")
	bob, bobConn := newClient("bob")

	s.handle(command{client: alice, name: "/msg", arg: "hello"})
	expect(t, aliceConn, ">: You must join the room first")

	s.handle(command{client: alice, name: "/join", arg: "go"})
	s.handle(command{client: bob, name: "/join", arg: "go"})
	received(aliceConn)
	received(bobConn)

	s.handle(command{client: alice, name: "/msg"})
	expect(t, aliceConn, ">: Message is required")

	s.handle(command{client: alice, name: "/msg", arg: "hello there"})
	expect(t, bobConn, ">: alice: hello there")
	expect(t, aliceConn) // sender doesn't get their own message

	if !reflect.DeepEqual(db.messages["go"], []string{"alice: hello there"}) {
		t.Fatalf("expected message to be saved, got %v", db.messages["go"])
	}
}

func TestQuit(t *testing.T) {
	s := newServer(newTestDB())
	alice, aliceConn := newClient("alice")
	bob, bobConn := newClient("bob")

	s.handle(command{client: alice, name: "/join", arg: "go"})
	s.handle(command{client: bob, name: "/join", arg: "go"})
	received(aliceConn)
	received(bobConn)

	s.handle(command{client: bob, name: "/quit"})
	expect(t, bobConn, ">: Goodbye bob")
	expect(t, aliceConn, ">: bob left the chatroom")
	if !bobConn.closed {
		t.Fatal("expected connection to be closed")
	}
}

func TestUnknownCommand(t *testing.T) {
	s := newServer(newTestDB())
	c, conn := newClient("alice")

	s.handle(command{client: c, name: "/dance"})
	expect(t, conn, "Error: Unknown command: /dance")
}
