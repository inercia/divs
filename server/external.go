package divs

import (
	"github.com/ccding/go-stun/stun"
	"github.com/huin/goupnp/dcps/internetgateway1"

	"fmt"
	"net"
	"strconv"
)

var ERR_COULD_NOT_OBTAIN = fmt.Errorf("Could not obtain a valid IP/port")

const STUN_SERVICE_ADDR = "stun.ekiga.net"
const STUN_SERVICE_PORT = 3478

type External struct {
	host string
	port int
}

// a new external address, with default values
func NewExternal(defaultHost string, defaultPort int) (*External, error) {
	return &External{
		host: defaultHost,
		port: defaultPort,
	}, nil
}

func (e *External) Obtain() (address string, err error) {
	var ip string
	var port int

	log.Info("Obtaining a valid external IP/port")

	if len(e.host) == 0 || e.port == 0 {
		ip, port, err = e.getExternalAddrWithUpnp()
		if err != nil {
			log.Debug("Trying with another method...")
			ip, port, err = e.getExternalAddrWithStun()
		}
		if err != nil {
			return "", err
		}
	}

	if len(e.host) != 0 {
		log.Debug("Forcing %s as external IP address\n", e.host)
		ip = e.host
	}

	if e.port != 0 {
		log.Debug("Using %d as external port\n", e.port)
		port = e.port
	}

	res := net.JoinHostPort(ip, strconv.Itoa(port))
	log.Info("Using %s as external IP", res)
	return res, nil
}

// try to obtain an external IP address and (optionally) port
func (e *External) getExternalAddrWithStun() (ip string, port int, err error) {
	err = nil
	var stunHost *stun.Host
	var nat int

	log.Debug("Using STUN for getting external IP from %s:%d...\n",
		STUN_SERVICE_ADDR, STUN_SERVICE_PORT)
	stun.SetServerHost(STUN_SERVICE_ADDR, STUN_SERVICE_PORT)
	nat, stunHost, err = stun.Discover()
	if err != nil {
		return "", 0, ERR_COULD_NOT_OBTAIN
	}
	if stunHost == nil {
		return "", 0, ERR_COULD_NOT_OBTAIN
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

// get an external IP and port with UpnP
func (e *External) getExternalAddrWithUpnp() (ip string, port int, err error) {
	log.Debug("Using UPnP for getting external IP/port")

	clients, errors, err := internetgateway1.NewWANPPPConnection1Clients()
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Got %d errors finding UPnP servers. %d UPnP servers discovered.\n",
		len(errors), len(clients))
	for i, e := range errors {
		log.Error("Error finding server #%d: %v\n", i+1, e)
	}
	if len(clients) == 0 {
		return "", 0, ERR_COULD_NOT_OBTAIN
	}

	for _, c := range clients {
		dev := &c.ServiceClient.RootDevice.Device
		srv := &c.ServiceClient.Service

		log.Debug(dev.FriendlyName, " :: ", srv.String())
		scpd, err := srv.RequestSCDP()
		if err != nil {
			log.Warning("  Error requesting service SCPD: %v\n", err)
		} else {
			log.Debug("  Available actions:")
			for _, action := range scpd.Actions {
				log.Debug("  * %s\n", action.Name)
				for _, arg := range action.Arguments {
					var varDesc string
					if stateVar := scpd.GetStateVariable(arg.RelatedStateVariable); stateVar != nil {
						varDesc = fmt.Sprintf(" (%s)", stateVar.DataType.Name)
					}
					log.Debug("    * [%s] %s%s\n", arg.Direction, arg.Name, varDesc)
				}
			}
		}

		if scpd == nil || scpd.GetAction("GetExternalIPAddress") != nil {
			ip, err := c.GetExternalIPAddress()
			log.Info("GetExternalIPAddress: ", ip, err)
		}
	}

	return ip, port, nil
}
