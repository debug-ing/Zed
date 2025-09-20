package tcp

import (
	"log"
	"net"
	"time"

	config "github.com/debug-ing/Zed/config/tcp"
	"github.com/debug-ing/Zed/internal/shared/server"

	"github.com/xtaci/smux"
)

type Server struct {
	conn net.Conn
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Disconnect() error {
	return s.conn.Close()
}

func (s *Server) Connect(cfg config.TcpServerConfig) {
	for {
		if err := s.connectOnce(cfg); err != nil {
			log.Println("reconnect after error:", err)
			time.Sleep(time.Second)
		}
	}
	// var err error
	// s.conn, err = net.Dial("tcp", cfg.ServerAddr)
	// if err != nil {
	// 	log.Fatal("dial:", err)
	// }
	// log.Println("connected to server tunnel")
	// ms, err := smux.Client(s.conn, nil)
	// if err != nil {
	// 	// _ = s.conn.Close()
	// 	_ = s.conn.Close()
	// 	return err
	// }
	// for {
	// 	fmt.Println("HH")
	// 	stream, err := ms.AcceptStream()
	// 	if err != nil {
	// 		log.Println("accept stream:", err)
	// 		// ms.Close()
	// 		// s.conn.Close()
	// 	}
	// 	fmt.Println("handled")
	// 	go server.HandleStream(stream)
	// }
}

func (s *Server) connectOnce(cfg config.TcpServerConfig) error {
	var err error
	s.conn, err = net.Dial("tcp", cfg.ServerAddr)
	if err != nil {
		return err
	}
	log.Println("connected to server tunnel")

	ms, err := smux.Client(s.conn, nil)
	if err != nil {
		_ = s.conn.Close()
		return err
	}

	for {
		stream, err := ms.AcceptStream()
		if err != nil {
			// IMPORTANT: smux session is dead
			ms.Close()
			s.conn.Close()
			return err // exit → reconnect
		}

		go server.HandleStream(stream)
	}
}
