package rendezvous

import (
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/nictuku/dht"
)

var ERR_CAN_NOT_DISCOVER_DHT = errors.New("Could not discover with DHT")

const DEFAULT_DHT_NODE = "213.239.195.138:40000"

//
const DISCOVERY_MIN_PEERS = 1

type DhtService struct {
	id            string
	ih            dht.InfoHash
	discoveryAddr string

	announcing  bool
	discovering bool
}

// Create a new DHT rendezvous service
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
		id:            id,
		ih:            ih,
		discoveryAddr: discoveryAddr,
	}
	return &d, nil
}

// Announce the service in the network
func (srv *DhtService) AnnounceAndDiscover(external string, discoveries chan string, localIPs *LocalIPs) error {
	log.Info("Starting WAN lookup with DHT")

	// Connect to the DHT network
	log.Debug("Connecting to DHT network...")
	_, port, _ := net.SplitHostPort(srv.discoveryAddr)
	portI, _ := strconv.Atoi(port)

	dhtConfig := dht.NewConfig()
	dhtConfig.Port = portI
	dhtService, err := dht.New(dhtConfig)
	if err != nil {
		log.Debug("Could not create the DHT node:", err)
		return ERR_CAN_NOT_DISCOVER_DHT
	}

	log.Debug("Adding DHT node %s...", DEFAULT_DHT_NODE)
	dhtService.AddNode(DEFAULT_DHT_NODE)

	go dhtService.DoDHT()
	go srv.peersDiscoveryWorker(dhtService, discoveries) // obtain peers from the DHT network

	return nil
}

func (srv *DhtService) Leave() error {
	return nil
}

// discover peers and send them to the discoveries channel
func (srv *DhtService) peersDiscoveryWorker(d *dht.DHT, discoveries chan string) {
	log.Debug("Waiting for possible peers...")
	lastPeersRequestTime := time.Now().Unix()
	for {
		select {
		case r := <-d.PeersRequestResults:
			for _, peers := range r {
				for _, x := range peers {
					// A DHT peer for our infohash was found. It
					// needs to be authenticated.
					address := dht.DecodePeerAddress(x)

					// TODO: we should do some challenge/response
					discoveries <- address
				}
			}
		case <-time.After(5 * time.Second):
			// nothing to do
		}

		if time.Now().Unix()-lastPeersRequestTime >= 5 {
			// Keeps requesting for the infohash. This is a no-op if the
			// DHT is satisfied with the number of peers it has found.
			d.PeersRequest(string(srv.ih), true)
			lastPeersRequestTime = time.Now().Unix()
		}
	}
}
