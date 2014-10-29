package divsd

// The top configuration structure for the DiVS daemon
type Config struct {
	Global   globalConfig
	Raft     raftConfig
	Discover discoverConfig
	Mdns     mdnsConfig
	Tun      tunConfig
}

// Global config
type globalConfig struct {
	Name   string
	Host   string
	Port   int
	Serial UUID
}

// RAFT
type raftConfig struct {
	DataPath string
	IsLeader bool
	Leader   string
}

// MDNS discovery
type mdnsConfig struct {
	Port   int
}

// DHT discovery
type discoverConfig struct {
	Port int
}

// NAT: TUN config
type tunConfig struct {
	NumReaders int
}

// Create a new DiVS daemon configuration
func NewConfig() (c *Config) {
	c = &Config{}
	return
}
