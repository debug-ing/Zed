package tcp

import (
	"log"
	"net"

	config "github.com/debug-ing/Zed/config/tcp"
	"github.com/debug-ing/Zed/internal/shared/server"

	"github.com/xtaci/smux"
)

func Server(cfg config.TcpServerConfig) {
	conn, err := net.Dial("tcp", cfg.ServerAddr)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("connected to server tunnel")
	ms, err := smux.Client(conn, nil)
	if err != nil {
		log.Fatal("smux client:", err)
	}
	for {
		stream, err := ms.AcceptStream()
		if err != nil {
			log.Println("accept stream:", err)
			return
		}
		go server.HandleStream(stream)
	}
}
