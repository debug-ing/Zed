package kcp

import (
	"log"
	"time"

	config "github.com/debug-ing/Zed/config/kcp"
	"github.com/debug-ing/Zed/internal/shared/server"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
)

type Server struct {
	con *kcp.UDPSession
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Disconnect() error {
	return s.con.Close()
}

func (s *Server) Connect(cfg config.KcpServerConfig) {
	backoff := time.Second
	for {
		if err := func() error {
			block, err := kcp.NewAESBlockCrypt([]byte(cfg.Key))
			if err != nil {
				return err
			}
			s.con, err = kcp.DialWithOptions(cfg.ServerAddr, block, cfg.DataShards, cfg.ParityShards)
			if err != nil {
				return err
			}

			log.Println("connected to server", cfg.ServerAddr)
			s.con.SetNoDelay(1, 10, 2, 1)
			s.con.SetWindowSize(1024, 1024)
			if cfg.Mtu != 0 {
				s.con.SetMtu(cfg.Mtu)
			}
			s.con.SetACKNoDelay(cfg.ACKNoDelay)
			ms, err := smux.Client(s.con, nil)
			if err != nil {
				_ = s.con.Close()
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
