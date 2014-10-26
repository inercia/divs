package rendezvous

import (
	"crypto/sha1"
	"crypto/sha256"
	"log"
	"time"
	"net"
	"strconv"
	"errors"

	"github.com/nictuku/dht"
)

var ERR_CAN_NOT_DISCOVER_DHT = errors.New("Could not discover with DHT")

const DEFAULT_DHT_NODE = "213.239.195.138:40000"

//
const DISCOVERY_MIN_PEERS = 1

type DhtService struct {
	id               string
	ih               dht.InfoHash
	discoveryAddr    string

	announcing  bool
	discovering bool

}

func NewDhtService(discoveryAddr string, id string) (*DhtService, error) {
	// infohash used for this wherez lookup. This should be somewhat hard to guess
	// but it's not exactly a secret.

	// SHA256 of the passphrase.
	h256 := sha256.New()
	h256.Write([]byte(id))
	h := h256.Sum(nil)

	// Assuming perfect rainbow databases, it's better if the infohash does not
	// give out too much about the passphrase. Take half of this hash, then
	// generate a SHA1 hash from it.
	h2 := h[0 : sha256.Size/2]

	// Mainline DHT uses sha1.
	h160 := sha1.New()
	h160.Write(h2)
	h3 := h160.Sum(nil)
	ih := dht.InfoHash(h3[:])

	d := DhtService{
		id:                    id,
		ih:                    ih,
		discoveryAddr:        discoveryAddr,
	}
	return &d, nil
}

func (this *DhtService) Announce(external string) error {
	return nil
}

func (this *DhtService) Discover(discoveries chan string) error {
	// Connect to the DHT network
	log.Println("Connecting to DHT network...")
	_, port, _ := net.SplitHostPort(this.discoveryAddr)
	portI, _ := strconv.Atoi(port)

	dhtConfig := dht.NewConfig()
	dhtConfig.Port = portI
	dhtService, err := dht.New(dhtConfig)
	if err != nil {
		log.Println("Could not create the DHT node:", err)
		return ERR_CAN_NOT_DISCOVER_DHT
	}

	log.Printf("Adding DHT node %s...", DEFAULT_DHT_NODE)
	dhtService.AddNode(DEFAULT_DHT_NODE)

	go dhtService.DoDHT()

	// obtins peers (that can authenticate) from the DHT network
	go func(d *dht.DHT) {
		log.Printf("Waiting for possible peers...")
		for r := range d.PeersRequestResults {
			for _, peers := range r {
				for _, x := range peers {
					// A DHT peer for our infohash was found. It
					// needs to be authenticated.
					address := dht.DecodePeerAddress(x)

					// TODO: we should do some challenge/response
					discoveries <- address
				}
			}
		}
	}(dhtService) // sends authenticated peers to channel c.

	for {
		// Keeps requesting for the infohash. This is a no-op if the
		// DHT is satisfied with the number of peers it has found.
		dhtService.PeersRequest(string(this.ih), true)
		time.Sleep(5 * time.Second)
	}

	return nil
}

func (this *DhtService) Leave() error {
	return nil
}
