package tcp

type TcpServerConfig struct {
	ServerAddr   string
	LocalTarget  string
	Key          string
	DataShards   int
	ParityShards int
}

type TcpAgentConfig struct {
	Listener     string
	Ports        []string
	Key          string
	DataShards   int
	ParityShards int
}
