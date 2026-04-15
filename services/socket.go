package services

import (
	"log"

	"github.com/zishang520/engine.io/v2/config"
	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/socket.io/v2/socket"
)

var SocketServer *socket.Server

func InitSocketServer() *socket.Server {
	opts := socket.DefaultServerOptions()
	
	// Configure CORS for engine.io
	eo := config.DefaultServerOptions()
	eo.SetCors(&types.Cors{
		Origin:      "*",
		Credentials: true,
	})
	// Allow v3/v4 clients (EIO=3, EIO=4)
	eo.SetAllowEIO3(true)
	
	opts.ServerOptions = *eo

	server := socket.NewServer(nil, opts)

	server.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		log.Println("connected:", client.Id())

		client.On("join", func(datas ...any) {
			log.Println("join", datas)
			client.Join("test_room")
			client.Emit("reply", "joined test_room")
		})

		client.On("disconnect", func(reasons ...any) {
			log.Println("disconnected:", client.Id(), reasons)
		})
	})

	SocketServer = server
	return server
}
