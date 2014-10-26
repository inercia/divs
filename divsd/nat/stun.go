package nat

import (
	"net"
	"github.com/ccding/go-stun/stun"
	"fmt"
	logging "github.com/op/go-logging"
	"strconv"
)

var log = logging.MustGetLogger("divs")

// could not obtain a NAT mapping with STUN
var ERR_COULD_NOT_OBTAIN_STUN = fmt.Errorf("Could not obtain a valid IP/port with STUN")

// default STUN server:port
const STUN_SERVICE_ADDRESS = "stun.ekiga.net:3478"

// try to obtain an external IP address and (optionally) port
func GetStun(defaultIp net.IP, defaultPort int) (net.IP, int, error) {
	var stunHost *stun.Host
	var nat int

	log.Debug("Using STUN for getting external IP from %s...\n",
		STUN_SERVICE_ADDRESS)
	sAddr, sPort, err := net.SplitHostPort(STUN_SERVICE_ADDRESS)
	if err != nil {
		return net.IP{}, 0, ERR_COULD_NOT_OBTAIN_STUN
	}
	sPortI, _ := strconv.Atoi(sPort)
	stun.SetServerHost(sAddr, sPortI)
	nat, stunHost, err = stun.Discover()
	if err != nil {
		return net.IP{}, 0, ERR_COULD_NOT_OBTAIN_STUN
	}
	if stunHost == nil {
		return net.IP{}, 0, ERR_COULD_NOT_OBTAIN_STUN
	}

	log.Debug("External endpoint calculated %s", stunHost.TransportAddr())
	var t string
	switch nat {
	case stun.NAT_ERROR:
		t = "test failed"
	case stun.NAT_UNKNOWN:
		t = "unexpected response from the STUN server"
	case stun.NAT_BLOCKED:
		t = "UDP is blocked"
	case stun.NAT_FULL:
		t = "Full cone NAT"
	case stun.NAT_SYMETRIC:
		t = "symetric NAT"
	case stun.NAT_RESTRICTED:
		t = "restricted NAT"
	case stun.NAT_PORT_RESTRICTED:
		t = "port restricted NAT"
	case stun.NAT_NONE:
		t = "not behind a NAT"
	case stun.NAT_SYMETRIC_UDP_FIREWALL:
		t = "symetric UDP firewall"
	}
	log.Debug("NAT type: %s.\n", t)

	return net.ParseIP(stunHost.Ip()), int(stunHost.Port()), nil
}
