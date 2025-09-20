package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	config "github.com/debug-ing/Zed/config/kcp"
	"github.com/debug-ing/Zed/internal/kcp"
)

// ---------------- Echo Servers ----------------
func udpEchoServer(wg *sync.WaitGroup, host string, port int) {
	defer wg.Done()
	addr := fmt.Sprintf("%s:%d", host, port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		fmt.Println("resolve error:", err)
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("udp listen error:", err)
		return
	}
	fmt.Printf("[UDP] listening on %s\n", addr)

	buf := make([]byte, 4096)
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("udp read error:", err)
			return
		}
		msg := append([]byte("echo:"), buf[:n]...)
		conn.WriteToUDP(msg, remote)
	}
}

func tcpEchoServer(wg *sync.WaitGroup, host string, port int) {
	defer wg.Done()
	addr := fmt.Sprintf("%s:%d", host, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("tcp listen error:", err)
		return
	}
	fmt.Printf("[TCP] listening on %s\n", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("tcp accept error:", err)
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 4096)
			for {
				n, err := c.Read(buf)
				if err != nil {
					return
				}
				c.Write(append([]byte("echo:"), buf[:n]...))
			}
		}(conn)
	}
}

// ---------------- Tunnel Test Clients ----------------
func testUDP(proxyHost string, proxyPort int) {
	addr := net.JoinHostPort(proxyHost, fmt.Sprintf("%d", proxyPort))
	conn, err := net.Dial("udp", addr)
	if err != nil {
		fmt.Println("udp dial error:", err)
		return
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(2 * time.Second))
	conn.Write([]byte("hello-udp"))

	buf := make([]byte, 4096)
	n, _, err := conn.(*net.UDPConn).ReadFromUDP(buf)
	if err != nil {
		fmt.Println("udp read error:", err)
		return
	}
	fmt.Println("[RESULT] UDP reply:", string(buf[:n]))
}

func testTCP(proxyHost string, proxyPort int) {
	addr := net.JoinHostPort(proxyHost, fmt.Sprintf("%d", proxyPort))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("tcp dial error:", err)
		return
	}
	defer conn.Close()

	conn.Write([]byte("hello-tcp"))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("tcp read error:", err)
		return
	}
	fmt.Println("[RESULT] TCP reply:", string(buf[:n]))
}

// ---------------- MAIN ----------------
func main() {
	var wg sync.WaitGroup
	// start echo servers
	wg.Add(2)
	go udpEchoServer(&wg, "127.0.0.1", 9000)
	go tcpEchoServer(&wg, "127.0.0.1", 8080)

	go func() {
		cfg := config.KcpAgentConfig{
			Listener: ":4000",
			Ports: []string{
				"8088:8080",
				"9001:9000",
			},
			Key:          "abcdefghijklmnop",
			DataShards:   10,
			ParityShards: 3,
		}
		kcp.Agent(cfg)
	}()
	//
	server := kcp.NewServer()
	//
	go func() {
		time.Sleep(2 * time.Second)
		cfg := config.KcpServerConfig{
			ServerAddr:   "127.0.0.1:4000",
			Key:          "abcdefghijklmnop",
			DataShards:   10,
			ParityShards: 3,
		}
		server.Connect(cfg)
	}()
	time.Sleep(15 * time.Second)
	testTCP("127.0.0.1", 8088)
	testUDP("127.0.0.1", 9001)
	time.Sleep(5 * time.Second)
	go func() {
		server.Disconnect()
		//
		cfg := config.KcpServerConfig{
			ServerAddr:   "127.0.0.1:4000",
			Key:          "abcdefghijklmnop",
			DataShards:   10,
			ParityShards: 3,
		}
		server.Connect(cfg)
	}()
	time.Sleep(30 * time.Second)
	testTCP("127.0.0.1", 8088)
	testUDP("127.0.0.1", 9001)
	fmt.Println("[DONE] killed tunnel processes")
}
