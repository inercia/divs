package rendezvous

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/hashicorp/mdns"
)

const PROTOCOL = "tcp"

type MdnsService struct {
	id            string
	fullId        string
	discoveryAddr string
	server        *mdns.Server
	entriesCh     chan *mdns.ServiceEntry
	announced     bool
	discovering   bool
}

// Create a new mDNS rendezvous service
func NewMdnsService(discoveryAddr string, id string) (*MdnsService, error) {
	m := MdnsService{
		id:          		id,
		fullId:      		fmt.Sprintf("_%s._%s", id, PROTOCOL),
		discoveryAddr:  	discoveryAddr,
		entriesCh:   		make(chan *mdns.ServiceEntry, 4),
		announced:   		false,
		discovering: 		false,
	}
	return &m, nil
}

// Announce the service in the network
func (srv *MdnsService) AnnounceAndDiscover(external string, discoveries chan string) error {
	// Setup our service export
	host, _ := os.Hostname()
	_, portStr, _ := net.SplitHostPort(external)
	port, _ := strconv.Atoi(portStr)

	log.Info("Announcing with mDNS service %s at %s:%d", srv.fullId, host, port)
	service := &mdns.MDNSService{
		Instance: host,
		Service:  srv.fullId,
		Port:     port,
	}
	service.Init()

	// Create the mDNS server, defer shutdown
	var err error
	srv.server, err = mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return err
	} else {
		srv.announced = true
	}

	go func() {
		for {
			entry, receiving := <-srv.entriesCh
			if !receiving {
				break
			}
			log.Debug("Got new peer: %v", entry)
			discoveries <- net.JoinHostPort(entry.Host, strconv.Itoa(entry.Port))
		}
	}()

	log.Info("Starting LAN lookup with mDNS for service %s", srv.fullId)
	mdns.Lookup(srv.fullId, srv.entriesCh)
	srv.discovering = true

	return nil
}

func (srv *MdnsService) Leave() error {
	if srv.announced {
		srv.server.Shutdown()
		srv.announced = false
	}
	if srv.discovering {
		close(srv.entriesCh)
		srv.discovering = false
	}

	return nil
}
