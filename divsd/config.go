package divsd

// The top config
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
	Serial string
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

// create a new config
func NewConfig() (c *Config) {
	c = &Config{}
	return
}
