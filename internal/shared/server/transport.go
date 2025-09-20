package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/debug-ing/Zed/internal/shared"
	"github.com/xtaci/smux"
)

type Protocol int

const (
	ProtocolUnknown Protocol = iota
	ProtocolTCP
	ProtocolUDP
)

func HandleStream(stream *smux.Stream) error {
	// if stream != nil {
	// 	fmt.Println("errro1")
	// 	defer stream.Close()
	// }
	buf := make([]byte, 4096)
	n, err := stream.Read(buf)
	if err != nil {
		return err
	}
	proto := detectProtocol(buf, n)
	switch proto {
	case ProtocolTCP:
		parts := bytes.SplitN(buf[1:n], []byte{0x00}, 2)
		if len(parts) < 2 {
			log.Println("invalid tcp header")
			return nil
		}
		dest := string(parts[0])
		firstPayload := parts[1]
		conn, err := net.Dial("tcp", dest)
		if err != nil {
			log.Println("dial target:", err)
			return nil
		}
		defer conn.Close()
		pr, pw := io.Pipe()
		go func() {
			if len(firstPayload) > 0 {
				pw.Write(firstPayload)
			}
			io.Copy(pw, stream)
			pw.Close()
		}()
		shared.PipeServer(conn, pr, stream)
	case ProtocolUDP:
		parts := bytes.SplitN(buf[1:n], []byte{0x00}, 2)
		if len(parts) < 2 {
			log.Println("invalid udp packet")
			return nil
		}

		dest := string(parts[0])
		payload := parts[1]

		udpAddr, err := net.ResolveUDPAddr("udp", dest)
		if err != nil {
			log.Println("resolve udp:", err)
			return nil
		}

		conn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			log.Println("dial udp:", err)
			return nil
		}
		defer conn.Close()
		_, _ = conn.Write(payload)
		reply := make([]byte, 4096)
		rn, _, err := conn.ReadFrom(reply)
		if err == nil {
			_, _ = stream.Write(reply[:rn])
		}
		return nil
	default:
		log.Println("unknown protocol:", buf[0])
		return nil
	}
	return nil
}

func detectProtocol(buf []byte, n int) Protocol {
	if n == 0 || len(buf) == 0 {
		return ProtocolUnknown
	}
	switch buf[0] {
	case 0x01:
		return ProtocolTCP
	case 0x02:
		return ProtocolUDP
	default:
		return ProtocolUnknown
	}
}

func parseHeader(data []byte) (string, []byte, error) {
	parts := bytes.SplitN(data, []byte{0x00}, 2)
	if len(parts) < 2 {
		return "", nil, fmt.Errorf("invalid header")
	}
	dest := string(parts[0])
	payload := parts[1]
	return dest, payload, nil
}
