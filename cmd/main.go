package main

import (
	"sync"

	"github.com/debug-ing/Zed/config"
	kcpConfig "github.com/debug-ing/Zed/config/kcp"
	tcpConfig "github.com/debug-ing/Zed/config/tcp"
	"github.com/debug-ing/Zed/internal/kcp"
	"github.com/debug-ing/Zed/internal/tcp"
	"github.com/xtaci/smux"
)

var SessMu sync.RWMutex
var MuxSession *smux.Session

func main() {

	cfg := config.LoadConfig()
	switch cfg.Mode.Mode {
	case "server":
		switch cfg.Mode.Type {
		case "tcp":
			server := tcp.NewServer()
			server.Connect(tcpConfig.TcpServerConfig{
				ServerAddr:   cfg.Server.Address,
				Key:          cfg.Server.Key,
				DataShards:   10,
				ParityShards: 3,
			})
		case "kcp":
			server := kcp.NewServer()
			server.Connect(kcpConfig.KcpServerConfig{
				ServerAddr:   cfg.Server.Address,
				Key:          cfg.Server.Key,
				DataShards:   10,
				ParityShards: 3,
				ACKNoDelay:   cfg.Kcp.ACKNoDelay,
				Mtu:          cfg.Kcp.Mtu,
				Internal:     cfg.Kcp.Internal,
			})
		}

	case "agent":
		switch cfg.Mode.Type {
		case "tcp":
			tcp.Agent(tcpConfig.TcpAgentConfig{
				Listener:     cfg.Agent.Address,
				Key:          cfg.Agent.Key,
				DataShards:   10,
				ParityShards: 3,
				Ports:        cfg.Agent.Ports,
			})
		case "kcp":
			kcp.Agent(kcpConfig.KcpAgentConfig{
				Listener:     cfg.Agent.Address,
				Key:          cfg.Agent.Key,
				DataShards:   10,
				ParityShards: 3,
				ACKNoDelay:   cfg.Kcp.ACKNoDelay,
				Mtu:          cfg.Kcp.Mtu,
				Internal:     cfg.Kcp.Internal,
				Ports:        cfg.Agent.Ports,
			})
		}
		for {

		}

	}
}
