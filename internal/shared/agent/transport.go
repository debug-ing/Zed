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
var once sync.Once

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
	// config := smux.DefaultConfig()
	// config.KeepAliveInterval = 2000
	// config.KeepAliveTimeout = 2000
	ms, err := smux.Server(con, nil)
	if err != nil {
		log.Println("smux server:", err)
		_ = con.Close()
		return true
	}
	stream, err := ms.OpenStream()
	if err != nil {
		fmt.Println("erre")
		_ = con.Close()
		return true
	}
	defer stream.Close()
	//

	// sessMu.Lock()
	// if muxSession != nil {
	// 	fmt.Println("sdf")
	// }
	// if muxSession != nil && !muxSession.IsClosed() {
	// 	sessMu.Unlock()
	// 	log.Println("another client already connected — rejecting new connection")
	// 	muxSession.Close() // close the newly created session
	// 	_ = con.Close()    // close raw connection too
	// 	return true
	// }
	muxSession = ms
	// sessMu.Unlock()

	//
	// once.Do(func() {
	for _, p := range ports {
		in, out := getInOut(p)
		// go forwardTCPListener(fmt.Sprintf(":%s", in), fmt.Sprintf("127.0.0.1:%s", out))
		go ForwardToClient(fmt.Sprintf(":%s", in), fmt.Sprintf("127.0.0.1:%s", out))
	}
	// })

	go func(ms *smux.Session) {
		<-ms.CloseChan()
		log.Println("smux session closed")
		sessMu.Lock()
		// muxSession = nil
		stream.Close()
		ms.Close()
		sessMu.Unlock()
	}(ms)

	return false
}

func forwardTCPListener(listenAddr, target string) {
	ln, _ := net.Listen("tcp", listenAddr)
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			sessMu.RLock()
			ms := muxSession
			sessMu.RUnlock()
			if ms == nil || ms.IsClosed() {
				log.Println("no active tunnel; rejecting connection from", c.RemoteAddr())
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
			go io.Copy(stream, c)
			io.Copy(c, stream)
		}(conn)
	}
}

func forwardUDPListener(listenAddr, target string) {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		log.Fatal("ResolveUDPAddr:", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal("ListenUDP:", err)
	}
	defer conn.Close()
	log.Println("UDP listening on", listenAddr)

	buf := make([]byte, 4096)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("UDP read error:", err)
			continue
		}

		// handle each packet in a goroutine
		go func(data []byte, addr *net.UDPAddr) {
			sessMu.RLock()
			ms := muxSession
			sessMu.RUnlock()

			if ms == nil || ms.IsClosed() {
				log.Println("no active tunnel, dropping UDP packet from", addr)
				return
			}

			stream, err := ms.OpenStream()
			if err != nil {
				log.Println("open stream for UDP:", err)
				return
			}
			defer stream.Close()

			pkt := []byte{0x02} // 0x02 = UDP
			pkt = append(pkt, []byte(target)...)
			pkt = append(pkt, 0x00)
			pkt = append(pkt, data...)

			_, err = stream.Write(pkt)
			if err != nil {
				log.Println("write stream for UDP:", err)
				return
			}

			reply := make([]byte, 4096)
			rn, err := stream.Read(reply)
			if err != nil {
				log.Println("read reply from stream:", err)
				return
			}

			_, err = conn.WriteToUDP(reply[:rn], addr)
			if err != nil {
				log.Println("write to UDP client:", err)
			}

		}(buf[:n], clientAddr)
	}
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
			stream, err := ms.AcceptStream()
			if err != nil {
				log.Println("open stream for udp:", err)
				// conn.Close()
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
	// go func() {
	// 	<-ms.CloseChan()
	// 	log.Println("Client disconnected, session closed")

	// 	sessMu.Lock()
	// 	if muxSession == ms {
	// 		muxSession = nil
	// 	}
	// 	sessMu.Unlock()
	// }()
	// go func() {
	for {
		c, err := tcpLn.Accept()
		if err != nil {
			log.Println("tcp accept:", err)
			continue
		}
		func(conn net.Conn) {
			fmt.Println("?")
			// defer conn.Close()
			sessMu.RLock()
			ms := muxSession
			sessMu.RUnlock()
			if ms == nil || ms.IsClosed() {
				log.Println("no active tunnel session; closing TCP conn from", conn.RemoteAddr())
				// conn.Close()
				return
			}
			stream, err := ms.AcceptStream()
			if err != nil {
				log.Println("open stream:", err)
				// conn.Close()
				return
			}
			defer stream.Close()
			pkt := getPacketTCP(target)
			stream.Write(pkt)
			shared.PipeAgent(conn, stream)
			// hdr := append([]byte{0x01}, []byte(target)...)
			// hdr = append(hdr, 0x00)
			// stream.Write(hdr)

			// go io.Copy(stream, conn)
			// io.Copy(conn, stream)
			// //
			// stream.Close()
			// // conn.Close()
		}(c)
	}
	// }()
}
