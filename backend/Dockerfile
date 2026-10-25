FROM golang:1.26.0
WORKDIR /chat
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./build/chatroom
EXPOSE 8888 8080
ENTRYPOINT ["/chat/build/chatroom"]