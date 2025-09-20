package agent

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/debug-ing/Zed/internal/shared"
	"github.com/xtaci/smux"
)

var sessMu sync.RWMutex
var muxSession *smux.Session

var (
	tcpListeners = make(map[string]net.Listener)
	udpConns     = make(map[string]*net.UDPConn)
	mu           sync.Mutex
)

func getTCPListener(listenAddr string) (net.Listener, error) {
	mu.Lock()
	defer mu.Unlock()

	if ln, ok := tcpListeners[listenAddr]; ok {
		return ln, nil
	}

	tcpLn, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	tcpListeners[listenAddr] = tcpLn
	log.Println("TCP public listening on", listenAddr)
	return tcpLn, nil
}

func getUDPConn(listenAddr string) (*net.UDPConn, error) {
	mu.Lock()
	defer mu.Unlock()

	if conn, ok := udpConns[listenAddr]; ok {
		return conn, nil
	}

	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	udpConns[listenAddr] = udpConn
	log.Println("UDP public listening on", listenAddr)
	return udpConn, nil
}

func Handle(con io.ReadWriteCloser, ports []string) bool {
	ms, err := smux.Server(con, nil)
	if err != nil {
		log.Println("smux server:", err)
		_ = con.Close()
		return true
	}
	for _, v := range ports {
		in, out := getInOut(v)
		go ForwardToClient(fmt.Sprintf(":%s", in), fmt.Sprintf("127.0.0.1:%s", out))
	}
	sessMu.Lock()
	if muxSession != nil {
		_ = muxSession.Close()
	}
	muxSession = ms
	sessMu.Unlock()
	go func(ms *smux.Session) {
		<-ms.CloseChan()
		log.Println("smux session closed")
		sessMu.Lock()
		muxSession = nil
		sessMu.Unlock()
	}(ms)
	return false
}

func getInOut(v string) (string, string) {
	in := strings.Split(v, ":")[0]
	out := strings.Split(v, ":")[1]
	return in, out
}

func getPacketUDP(target string, buf []byte, n int) []byte {
	packet := append([]byte{0x02}, []byte(target)...)
	packet = append(packet, 0x00)
	packet = append(packet, buf[:n]...)
	return packet
}
func getPacketTCP(target string) []byte {
	packet := append([]byte{0x01}, []byte(target)...)
	packet = append(packet, 0x00)
	return packet
}

func ForwardToClient(listenAddr, target string) {
	tcpLn, err := getTCPListener(listenAddr)
	if err != nil {
		log.Fatal("tcp listen:", err)
	}
	defer tcpLn.Close()

	udpConn, err := getUDPConn(listenAddr)
	if err != nil {
		log.Fatal("udp listen:", err)
	}
	defer udpConn.Close()
	go func() {
		buf := make([]byte, 4096)
		for {
			n, clientAddr, err := udpConn.ReadFromUDP(buf)
			if err != nil {
				log.Println("udp read:", err)
				continue
			}
			sessMu.Lock()
			ms := muxSession
			sessMu.Unlock()
			if ms == nil {
				log.Println("no active tunnel for UDP")
				continue
			}
			stream, err := ms.OpenStream()
			if err != nil {
				log.Println("open stream for udp:", err)
				continue
			}
			// pkt := append([]byte{0x02}, []byte(target)...)
			// pkt = append(pkt, 0x00)
			// pkt = append(pkt, buf[:n]...)
			pkt := getPacketUDP(target, buf, n)
			_, err = stream.Write(pkt)
			if err != nil {
				log.Println("write stream:", err)
				stream.Close()
				continue
			}

			reply := make([]byte, 4096)
			rn, err := stream.Read(reply)
			if err == nil {
				udpConn.WriteToUDP(reply[:rn], clientAddr)
			} else {
				log.Println("read from stream:", err)
			}
			stream.Close()
		}
	}()
	for {
		c, err := tcpLn.Accept()
		if err != nil {
			log.Println("tcp accept:", err)
			continue
		}
		go func(conn net.Conn) {
			// defer conn.Close()
			sessMu.RLock()
			ms := muxSession
			sessMu.RUnlock()
			if ms == nil || ms.IsClosed() {
				log.Println("no active tunnel session; closing TCP conn from", conn.RemoteAddr())
				return
			}
			stream, err := ms.OpenStream()
			if err != nil {
				log.Println("open stream:", err)
				return
			}
			defer stream.Close()
			pkt := getPacketTCP(target)
			stream.Write(pkt)
			shared.PipeAgent(conn, stream)
		}(c)
	}
}
