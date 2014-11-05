// The rendezvous package is responsible for starting the discovery service
// The discovery service must:
// * announce in the network the presence if this node
// * lookup other nodes that are using the same service
package rendezvous

import (
	"net"
	"fmt"
	logging "github.com/op/go-logging"
	"github.com/inercia/divs/divsd/nat"

)

const LOG_MODULE = "divs"

var log = logging.MustGetLogger(LOG_MODULE)


type LocalIPs map[string]bool

func NewLocalIps() *LocalIPs {
	localIPs := make(LocalIPs)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Debug("local IPs error: %s", err)
	} else {
		for _, a := range addrs {
			ip, ipnet, err := net.ParseCIDR(a.String())
			if err != nil {
				log.Debug("local IPs error: %s", err)
			} else {
				log.Debug("Found local IP %s [%s]", ip, ipnet)
				localIPs[ip.String()] = true
			}
		}
	}

	return &localIPs
}

func (localIps LocalIPs) IsLocal(ip string) bool {
	_, found := localIps[ip]
	return found
}

////////////////////////////////////////////////////////////////////////////////

// A rendezvous service: something we can use for joining a groups of nodes that
// also provide a service, or for announcing ourselves...
type RendezvousService interface {
	// Announce the service (publishing with the external address provided) and
	// discover other nodes for the same service
	AnnounceAndDiscover(string, chan string, *LocalIPs) error

	// Leave the rendezvous
	Leave() error
}

////////////////////////////////////////////////////////////////////////////////

func Start(serviceId string, bindIp string, dhtPort int, externalAddr string, discoveredChan chan string) {
	localIPs := NewLocalIps()

	// create the MDNS service
	go func() {
		mdnsService, err := NewMdnsService("", serviceId)
		if err != nil {
			log.Error("Could not start the mDNS service")
		} else {
			mdnsService.AnnounceAndDiscover(externalAddr, discoveredChan, localIPs)
		}
	}()

	// create the DHT service by previously obtaining an external TCP address
	go func() {
		defaultAddr := fmt.Sprintf("%s:%d", bindIp, dhtPort)
		dhtAddr, err := nat.NewExternalTCPAddr(defaultAddr)
		if err != nil {
			log.Error("Could not obtain an external port for the DHT service")
		} else {
			dhtService, err := NewDhtService(dhtAddr.String(), serviceId)
			if err != nil {
				log.Error("Could not start the DHT service")
			} else {
				dhtService.AnnounceAndDiscover(externalAddr, discoveredChan, localIPs)
			}
		}
	}()
}
