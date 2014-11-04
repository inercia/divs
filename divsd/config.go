package divsd

// The top configuration structure for the DiVS daemon
type Config struct {
	Global   globalConfig
	Discover discoverConfig
	Mdns     mdnsConfig
	Tun      tunConfig
}

// Global config
type globalConfig struct {
	Name   string
	Host   string
	Port   int
	BindIP string
	Serial UUID
}

// MDNS discovery
type mdnsConfig struct {
	Port int
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
