package divsd

import (
	"github.com/inercia/divs/divsd/nat"
	"net"
	"strconv"
)

type resolversFunc func(net.IP, int) (net.IP, int, error)

var resolvers = []resolversFunc{
	nat.GetUpnp,
	nat.GetStun,
}

func getExternalAddr(defaultIp net.IP, defaultPort int) (net.IP, int, error) {
	for _, resolver := range resolvers {
		if ip, port, err := resolver(defaultIp, defaultPort); err == nil {
			return ip, port, nil
		}
	}

	log.Debug("Returning default external binding")
	return defaultIp, defaultPort, nil
}

// Obtain a new external TCP address, providing a default value
func NewExternalTCP(tcpAddr net.TCPAddr) (net.TCPAddr, error) {
	log.Info("Obtaining a valid external TCP IP/port")
	if ip, port, err := getExternalAddr(tcpAddr.IP, tcpAddr.Port); err == nil {
		return net.TCPAddr{IP: ip, Port: port}, nil
	}
	return net.TCPAddr{}, ERR_COULD_NOT_OBTAIN_NAT
}

// Obtain a new external TCP address, providing the default value as a string
func NewExternalTCPAddr(addr string) (net.TCPAddr, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Error("Could not parse address %s", addr)
		return net.TCPAddr{}, err
	}
	portI, _ := strconv.Atoi(port)
	tcpAddr := net.TCPAddr{net.ParseIP(host), portI, ""}
	return NewExternalTCP(tcpAddr)
}

// Obtain a new external UDP address, providing a default value
func NewExternalUDP(udpAddr net.UDPAddr) (net.UDPAddr, error) {
	log.Info("Obtaining a valid external UDP IP/port")
	if ip, port, err := getExternalAddr(udpAddr.IP, udpAddr.Port); err == nil {
		return net.UDPAddr{IP: ip, Port: port}, nil
	}
	return net.UDPAddr{}, ERR_COULD_NOT_OBTAIN_NAT
}

// Obtain a new external UDP address, providing the default value as a string
func NewExternalUDPAddr(addr string) (net.UDPAddr, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Error("Could not parse address %s", addr)
		return net.UDPAddr{}, err
	}
	portI, _ := strconv.Atoi(port)
	udpAddr := net.UDPAddr{net.ParseIP(host), portI, ""}
	return NewExternalUDP(udpAddr)
}

// Obtain a new external address, returning a UDP or TCP address depending on
// the default value provided.
func NewExternal(addr interface{}) (interface{}, error) {
	switch addr.(type) {
	case net.TCPAddr:
		return NewExternalTCP(addr.(net.TCPAddr))
	case net.UDPAddr:
		return NewExternalUDP(addr.(net.UDPAddr))
	}
	return nil, ERR_COULD_NOT_OBTAIN_NAT
}
