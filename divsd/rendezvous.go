package divsd

import (
	"fmt"
	"github.com/inercia/divs/divsd/rendezvous"

)

// A rendezvous service: something we can use for joining a groups of nodes that
// also provide a service, or for announcing ourselves...
type RendezvousService interface {
	// Announce the service (publishing with the external address provided) and
	// discover other nodes for the same service
	AnnounceAndDiscover(string, chan string) error

	// Leave the rendezvous
	Leave() error
}

func startRendezVous(config *Config, externalAddr string, discoveredChan chan string) {
	serviceId := config.Global.Serial.ToHex()

	// create the MDNS service
	go func() {
		mdnsService, err := rendezvous.NewMdnsService("", serviceId)
		if err != nil {
			log.Error("Could not start the mDNS service")
		} else {
			mdnsService.AnnounceAndDiscover(externalAddr, discoveredChan)
		}
	}()

	// create the DHT service by previously obtaining an external TCP address
	go func() {
		defaultAddr := fmt.Sprintf("0.0.0.0:%d", config.Discover.Port)
		dhtAddr, err := NewExternalTCPAddr(defaultAddr)
		if err != nil {
			log.Error("Could not obtain an external port for the DHT service")
		} else {
			dhtService, err := rendezvous.NewDhtService(dhtAddr.String(), serviceId)
			if err != nil {
				log.Error("Could not start the DHT service")
			} else {
				dhtService.AnnounceAndDiscover(externalAddr, discoveredChan)
			}
		}
	}()
}
