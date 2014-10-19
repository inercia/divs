package rendezvous

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/armon/mdns"
)

const PROTOCOL = "tcp"

type MdnsService struct {
	id          string
	fullId      string
	server      *mdns.Server
	entriesCh   chan *mdns.ServiceEntry
	announced   bool
	discovering bool
}

func NewMdnsService(id string) (*MdnsService, error) {
	m := MdnsService{
		id:          id,
		fullId:      fmt.Sprintf("_%s._%s", id, PROTOCOL),
		entriesCh:   make(chan *mdns.ServiceEntry, 4),
		announced:   false,
		discovering: false,
	}
	return &m, nil
}

func (this *MdnsService) Announce(external string) error {
	// Setup our service export
	host, _ := os.Hostname()
	_, portStr, _ := net.SplitHostPort(external)
	port, _ := strconv.Atoi(portStr)

	service := &mdns.MDNSService{
		Instance: host,
		Service:  this.fullId,
		Port:     port,
		Info:     "My awesome service",
	}
	service.Init()

	// Create the mDNS server, defer shutdown
	var err error
	this.server, err = mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return err
	} else {
		this.announced = true
	}

	return nil
}

func (this *MdnsService) Discover(discoveries chan string) error {
	go func() {
		for entry := range this.entriesCh {
			// TODO: send to the "discoveries" channel
			fmt.Printf("Got new entry: %v\n", entry)
		}
	}()

	mdns.Lookup(this.fullId, this.entriesCh)
	this.discovering = true
	return nil
}

func (this *MdnsService) Leave() error {
	if this.announced {
		this.server.Shutdown()
	}
	if this.discovering {
		close(this.entriesCh)
	}

	return nil
}
