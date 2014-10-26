package divsd

import (
	"net"
	"github.com/inercia/divs/divsd/nat"
	"strconv"
)

type resolversFunc func(net.IP, int) (net.IP, int, error)

var resolvers = []resolversFunc {
	nat.GetUpnp,
	nat.GetStun,
}

func getExternalAddr(defaultIp net.IP, defaultPort int) (net.IP, int, error) {
	log.Info("Obtaining a valid external IP/port")
	for _, resolver := range resolvers {
		if gIp, gPort, err := resolver(defaultIp, defaultPort); err == nil {
			return gIp, gPort, nil
		}
	}
	return defaultIp, defaultPort, nil
}

// Obtain a new external TCP address
func NewExternalTCP(tcpAddr net.TCPAddr) (net.TCPAddr, error) {
	ip, port, err := getExternalAddr(tcpAddr.IP, tcpAddr.Port)
	if err != nil {
		return net.TCPAddr{IP:ip, Port:port}, nil
	}
	return net.TCPAddr{}, nil
}

func NewExternalTCPAddr(addr string) (net.TCPAddr, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return net.TCPAddr{}, err
	}
	portI, _ := strconv.Atoi(port)
	tcpAddr := net.TCPAddr{net.ParseIP(host), portI, ""}
	return NewExternalTCP(tcpAddr)
}

// Obtain a new external UDP address
func NewExternalUDP(udpAddr net.UDPAddr) (net.UDPAddr, error) {
	ip, port, err := getExternalAddr(udpAddr.IP, udpAddr.Port)
	if err != nil {
		return net.UDPAddr{IP:ip, Port:port}, nil
	}
	return net.UDPAddr{}, nil
}

func NewExternalUDPAddr(addr string) (net.UDPAddr, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return net.UDPAddr{}, err
	}
	portI, _ := strconv.Atoi(port)
	udpAddr := net.UDPAddr{net.ParseIP(host), portI, ""}
	return NewExternalUDP(udpAddr)
}

// a new external address, with default values
func NewExternal(addr interface{}) (interface{}, error) {
	switch addr.(type) {
	case net.TCPAddr:
		return NewExternalTCP(addr.(net.TCPAddr))
	case net.UDPAddr:
		return NewExternalUDP(addr.(net.UDPAddr))
	}
	return nil, nil
}

