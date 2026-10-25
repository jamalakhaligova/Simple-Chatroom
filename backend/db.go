package main

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type store struct {
	db *sql.DB
}

func newStore(url string) (*store, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS rooms (
			name TEXT PRIMARY KEY
		);
		CREATE TABLE IF NOT EXISTS messages (
			id         SERIAL PRIMARY KEY,
			room       TEXT NOT NULL REFERENCES rooms(name),
			username   TEXT NOT NULL,
			body       TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`)
	if err != nil {
		return nil, err
	}

	return &store{db: db}, nil
}

func (s *store) saveRoom(name string) error {
	_, err := s.db.Exec("INSERT INTO rooms (name) VALUES ($1) ON CONFLICT DO NOTHING", name)
	return err
}

func (s *store) loadRooms() ([]string, error) {
	rows, err := s.db.Query("SELECT name FROM rooms")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		rooms = append(rooms, name)
	}
	return rooms, rows.Err()
}

func (s *store) saveMessage(room, username, body string) error {
	_, err := s.db.Exec("INSERT INTO messages (room, username, body) VALUES ($1, $2, $3)", room, username, body)
	return err
}

// lastMessages returns the latest n messages of a room, oldest first.
func (s *store) lastMessages(room string, n int) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT username, body FROM (
			SELECT id, username, body FROM messages WHERE room = $1 ORDER BY id DESC LIMIT $2
		) AS recent ORDER BY id`, room, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []string
	for rows.Next() {
		var username, body string
		if err := rows.Scan(&username, &body); err != nil {
			return nil, err
		}
		msgs = append(msgs, username+": "+body)
	}
	return msgs, rows.Err()
}
