FROM golang:1.25.0
WORKDIR /chat
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./build/chatroom
EXPOSE 8888
ENTRYPOINT ["/chat/build/chatroom"]