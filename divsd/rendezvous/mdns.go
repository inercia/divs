package rendezvous

import (
	"fmt"
	"net"
	"strconv"

	"github.com/oleksandr/bonjour"
)

const PROTOCOL = "tcp"


////////////////////////////////////////////////////////////////////////////////

type MdnsService struct {
	id               string
	fullId           string
	resolver         *bonjour.Resolver
	registerStopChan chan<- bool
	discoveryAddr    string
	discoveriesChan  chan *bonjour.ServiceEntry
	discovering      bool
	announced        bool
}

// Create a new mDNS rendezvous service
func NewMdnsService(discoveryAddr string, id string) (*MdnsService, error) {
	resolver, err := bonjour.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	m := MdnsService{
		id:              id,
		fullId:          fmt.Sprintf("_%s._%s", id, PROTOCOL),
		resolver:        resolver,
		discoveryAddr:   discoveryAddr,
		discoveriesChan: make(chan *bonjour.ServiceEntry),
		announced:       false,
		discovering:     false,
	}
	return &m, nil
}

// Announce the service in the network
func (srv *MdnsService) AnnounceAndDiscover(external string, discoveries chan string, localIPs *LocalIPs) error {
	// Setup our service export
	_, portStr, _ := net.SplitHostPort(external)
	port, _ := strconv.Atoi(portStr)

	// Run registration
	log.Info("Announcing with mDNS service %s at :%d", srv.fullId, port)
	stopChan, err := bonjour.Register("The incredible foo service",
		srv.fullId, "", port,
		[]string{"txtv=1", "app=divs"}, nil)
	if err != nil {
		log.Error("Could not register with mDNS: %s", err)
		return err
	} else {
		srv.announced = true
		srv.registerStopChan = stopChan
	}

	// Create the mDNS server, defer shutdown
	go func(results chan *bonjour.ServiceEntry) {
		for entry := range results {
			var entryHostName string
			if entry.AddrIPv4 == nil || entry.AddrIPv4.IsUnspecified() {
				entryHostName = entry.HostName
			} else {
				entryHostName = entry.AddrIPv4.String()
			}

			entryAddr := fmt.Sprintf("%s:%d", entryHostName, entry.Port)
			log.Debug("Located a peer with mDNS: %s", entryAddr)

			// check if we have discovered ourselves...
			if localIPs.IsLocal(entryHostName) && entry.Port == port {
				log.Debug("... skipped: it was this node (%s:%d)", entryHostName, port)
			} else {
				discoveries <- entryAddr
			}
		}
	}(srv.discoveriesChan)

	log.Info("Starting LAN lookup with mDNS for service %s", srv.fullId)
	err = srv.resolver.Browse(srv.fullId, "local.", srv.discoveriesChan)
	if err != nil {
		log.Error("Could not start mDNS discovery: %s", err)
		return err
	} else {
		srv.discovering = true
	}

	return nil
}

// Stop announcing the service and stop discovering peers...
func (srv *MdnsService) Leave() error {
	if srv.announced {
		srv.registerStopChan <- true
		srv.announced = false
	}
	if srv.discovering {
		close(srv.discoveriesChan)
		srv.discovering = false
	}

	return nil
}
