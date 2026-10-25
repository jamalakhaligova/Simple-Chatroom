### Simple Chatroom in TCP

After connecting to chatroom via `telnet localhost 8888`, following endpoints are available:

- `/user <username>`: new user joins the chat
- `/rooms`: displays available chatrooms
- `/join <chatroom>` : user joins listed or new chatroom
- `/msg <message>` : user sends a message and it is broadcasted to everyone
- `/quit` : user leaves the chatroom and closes connection

#### Project layout

```
backend/    Go TCP chat server + WebSocket bridge for the browser
frontend/   index.html served by nginx, which proxies /ws to the backend
docker-compose.yml
```

#### Run with Docker Compose

```
docker compose up --build
```

- Web UI: http://localhost:8080
- Telnet: `telnet localhost 8888`

Web and telnet users can talk to each other.

#### Tests

```
cd backend && go test ./...
```

The tests use an in-memory store, so no database is needed.

#### Database
Rooms and messages are stored in PostgreSQL (the `db` service, data kept in the `db-data` volume). When you join a room, the last 20 messages are shown. The backend reads the connection string from `DATABASE_URL`.
