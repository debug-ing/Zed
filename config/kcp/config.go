package kcp

type KcpServerConfig struct {
	ServerAddr   string
	LocalTarget  string
	Key          string
	DataShards   int
	ParityShards int
	ACKNoDelay   bool
	Mtu          int
	Internal     int
}

type KcpAgentConfig struct {
	Listener     string
	Ports        []string
	Key          string
	DataShards   int
	ParityShards int
	ACKNoDelay   bool
	Mtu          int
	Internal     int
}
