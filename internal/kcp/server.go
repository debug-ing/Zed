package kcp

import (
	"log"
	"time"

	config "github.com/debug-ing/Zed/config/kcp"
	"github.com/debug-ing/Zed/internal/shared/server"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
)

func Server(cfg config.KcpServerConfig) {
	backoff := time.Second
	for {
		if err := func() error {
			block, err := kcp.NewAESBlockCrypt([]byte(cfg.Key))
			if err != nil {
				return err
			}
			session, err := kcp.DialWithOptions(cfg.ServerAddr, block, cfg.DataShards, cfg.ParityShards)
			if err != nil {
				return err
			}
			log.Println("connected to server", cfg.ServerAddr)
			session.SetNoDelay(1, 10, 2, 1)
			session.SetWindowSize(1024, 1024)
			if cfg.Mtu != 0 {
				session.SetMtu(cfg.Mtu)
			}
			session.SetACKNoDelay(cfg.ACKNoDelay)
			ms, err := smux.Client(session, nil)
			if err != nil {
				_ = session.Close()
				return err
			}
			defer ms.Close()
			for {
				stream, err := ms.AcceptStream()
				if err != nil {
					return err
				}
				go server.HandleStream(stream)
			}
		}(); err != nil {

		}
		log.Println("reconnecting in", backoff)
		time.Sleep(backoff)
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}
