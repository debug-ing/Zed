package tcp

import (
	"log"
	"net"

	config "github.com/debug-ing/Zed/config/tcp"
	"github.com/debug-ing/Zed/internal/shared/agent"
)

func Agent(cfg config.TcpAgentConfig) {
	ln, err := net.Listen("tcp", cfg.Listener)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Tunnel TCP server listening on ", cfg.Listener)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println("accept:", err)
				conn.Close()
				continue
			}
			log.Println("new tunnel conn:", conn.RemoteAddr())
			status := agent.Handle(conn, cfg.Ports)
			if status {
				continue
			}
		}
	}()

}
