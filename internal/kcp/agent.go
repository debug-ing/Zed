package kcp

import (
	"log"

	config "github.com/debug-ing/Zed/config/kcp"
	"github.com/debug-ing/Zed/internal/shared/agent"
	"github.com/xtaci/kcp-go/v5"
)

func Agent(cfg config.KcpAgentConfig) {
	block, err := kcp.NewAESBlockCrypt([]byte(cfg.Key))
	if err != nil {
		log.Fatal(err)
	}
	lis, err := kcp.ListenWithOptions(cfg.Listener, block, cfg.DataShards, cfg.ParityShards)
	if err != nil {
		log.Fatal("kcp listen:", err)
	}
	log.Println("KCP server listening on", cfg.Listener)
	lis.SetReadBuffer(512 * 1024)  //2048 //  4 << 20
	lis.SetWriteBuffer(512 * 1024) //2048 //  4 << 20
	go func() {
		for {
			session, err := lis.AcceptKCP()
			if err != nil {
				log.Println("session kcp:", err)
				continue
			}
			log.Println("KCP session arrived from", session.RemoteAddr())
			session.SetNoDelay(1, 20, 2, 1)
			session.SetWindowSize(1024, 1024)
			if cfg.Mtu != 0 {
				session.SetMtu(cfg.Mtu)
			}
			session.SetACKNoDelay(cfg.ACKNoDelay) // true
			status := agent.Handle(session, cfg.Ports)
			if status {
				continue
			}
		}
	}()
}
