package divsd

// The config
type Config struct {
	Global   globalConfig
	Raft     raftConfig
	Discover discoverConfig
	Mdns     mdnsConfig
	Tun      tunConfig
}

type globalConfig struct {
	Name   string
	Host   string
	Port   int
	Serial string
}

type mdnsConfig struct {
	Port   int
}

type raftConfig struct {
	DataPath string
	IsLeader bool
	Leader   string
}

type discoverConfig struct {
	Port int
}

type tunConfig struct {
	NumReaders int
}

// create a new config
func NewConfig() (c *Config) {
	c = &Config{}
	return
}
