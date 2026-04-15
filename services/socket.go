package services

import (
	"log"

	socketio "github.com/googollee/go-socket.io"
)

var SocketServer *socketio.Server

func InitSocketServer() *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Println("connected:", s.ID())
		return nil
	})

	server.OnEvent("/", "join", func(s socketio.Conn, msg string) {
		s.Join("test_room")
		s.Emit("reply", "joined test_room")
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("closed", reason)
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Printf("socketio listen error: %s\n", err)
		}
	}()

	SocketServer = server
	return server
}
