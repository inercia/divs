package divs

import (
	"github.com/ccding/go-stun/stun"
	"github.com/huin/goupnp/dcps/internetgateway1"

	"fmt"
	"net"
	"strconv"
)

var ERR_COULD_NOT_OBTAIN_STUN = fmt.Errorf("Could not obtain a valid IP/port with STUN")

const STUN_SERVICE_ADDR = "stun.ekiga.net"
const STUN_SERVICE_PORT = 3478

// try to obtain an external IP address and (optionally) port
func (e *External) Get() (ip string, port int, err error) {
	err = nil
	var stunHost *stun.Host
	var nat int

	log.Debug("Using STUN for getting external IP from %s:%d...\n",
		STUN_SERVICE_ADDR, STUN_SERVICE_PORT)
	stun.SetServerHost(STUN_SERVICE_ADDR, STUN_SERVICE_PORT)
	nat, stunHost, err = stun.Discover()
	if err != nil {
		return "", 0, ERR_COULD_NOT_OBTAIN_STUN
	}
	if stunHost == nil {
		return "", 0, ERR_COULD_NOT_OBTAIN_STUN
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

	return stunHost.Ip(), int(stunHost.Port()), nil
}
