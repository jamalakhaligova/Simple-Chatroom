### Simple Chatroom in TCP

After connecting to chatroom via `telnet 8888`, following endpoints are available:

- `/user <username>`: new user joins the chat
- `/rooms`: displays available chatrooms
- `/join <chatroom>` : user joins listed or new chatroom
- `/msg <message>` : user sends a message and it is broadcasted to everyone
- `/quit` : user leaves the chatroom and closes connection

#### Run Docker image using port mapping
After building the image, for example `docker build . -t chatroom`, do port mapping `docker run -p 8888:8888 chatroom`