package tcp

import (
	"fmt"
	"log"
	"net"

	config "github.com/debug-ing/Zed/config/tcp"
	"github.com/debug-ing/Zed/internal/shared/server"

	"github.com/xtaci/smux"
)

type Server struct {
	conn net.Conn
	ms   *smux.Session
	st   *smux.Stream
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Disconnect() error {
	s.st.Close()
	s.ms.Close()
	return s.conn.Close()
}

func (s *Server) Connect(cfg config.TcpServerConfig) {
	// for {
	// if err := s.connectOnce(cfg); err != nil {
	// 	log.Println("reconnect after error:", err)
	// 	time.Sleep(time.Second)
	// }
	var err error
	s.conn, err = net.Dial("tcp", cfg.ServerAddr)
	if err != nil {
		log.Println("reconnect after error:", err)
	}
	defer s.conn.Close()

	s.ms, err = smux.Client(s.conn, nil)
	if err != nil {
		_ = s.conn.Close()
		log.Println("reconnect after error:", err)
	}

	// defer stream.Close()
	for {
		s.st, err = s.ms.OpenStream()
		if err != nil {
			fmt.Println("ere")
		}
		defer s.st.Close()
		if err := server.HandleStream(s.st); err != nil {
			log.Println("reconnect after error:", err)
			s.st.Close()
			s.ms.Close()
			break
		}
	}
}
