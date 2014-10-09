package server

import (
	"fmt"
	"github.com/inercia/wherez/discover"
	"net"
	"sync"
	"time"
)

// the peers manager is responsible for
// - establishing and keeping connections to peers
// - sending/receiving data to/from peers
type PeersManager struct {
	config *Config

	discoveredChan chan discover.Peer
	peers          map[string]*Peer
	peerMux        sync.Mutex
	newPeer        chan bool // we'll send to this channel when we discover a new peer

	external       *External
	discoverer     *discover.Discoverer
	discovererStop chan int

	devManager *DevManager

	ExternalAddr string

	sync.RWMutex
}

// create a new peers manager
func NewPeersManager(config *Config, dm *DevManager) (d *PeersManager, err error) {
	discoverer, derr := discover.NewDiscoverer(config.Discover.Port,
		config.Global.Port, []byte(config.Global.Serial))
	if derr != nil {
		return nil, derr
	}

	log.Info("Creating new peers manager")
	d = &PeersManager{
		config:         config,
		devManager:     dm,
		peers:          make(map[string]*Peer),
		newPeer:        make(chan bool),
		discoverer:     discoverer,
		discovererStop: make(chan int),
		external:       NewExternal(config.Global.Host, config.Global.Port),
	}
	return d, nil
}

// start the peers manager
func (p *PeersManager) Start() (err error) {
	// calculate the external IP and port other peers will use for connecting to us
	if p.ExternalAddr, err = p.external.Obtain(); err != nil {
		return err
	}

	// perform peers discovery in another goroutine...
	go func() {
		for {
			log.Info("Starting peers discovery for id=%s...", p.config.Global.Serial)
			p.discoverer.FindPeers(1)

			select {
			case np := <-p.discoverer.DiscoveredPeers:
				log.Debug("New peer discovered: %s", np.Addr)

				// create the new peer and insert it in the map
				if newPeer, npErr := NewPeer(np.Addr); npErr != nil {
					log.Warning("could not create new peer:", npErr)
				} else {
					p.addPeer(newPeer)
				}

			case <-p.discovererStop:
				log.Debug("Exiting discovery loop")
				break
			}

		}
	}()

	return nil
}

// wait some time for some peers
func (d *PeersManager) StopDiscovery() (err error) {
	log.Debug("Signaling stop for discovery")
	close(d.discovererStop)
	return nil
}

// Add a new peer and notify any waiter
func (p *PeersManager) addPeer(newPeer *Peer) {
	p.peerMux.Lock()
	p.peers[newPeer.Id] = newPeer
	p.peerMux.Unlock()
	p.newPeer <- true
}

// wait some time for some peers
func (d *PeersManager) WaitForPeersTime(seconds time.Duration) (err error) {
	log.Info("Waiting for %d seconds for peers to join...", seconds)
	select {
	case <-d.newPeer:
		err = nil
		break
	case <-time.After(seconds * time.Second):
		err = fmt.Errorf("no valid peers found in %d seconds", seconds)
	}
	return err
}

// wait for some peers
func (p *PeersManager) WaitForPeers() error {
	log.Info("Waiting forever for new peers to join...")
	<-p.newPeer
	return nil
}

// adds a packet for a destination
func (p *PeersManager) SendTo(packet []byte, peerId string) (n int, err error) {
	peer, found := p.peers[peerId]
	if !found {
		return 0, fmt.Errorf("no peer found when sending data")
	}

	n, err = peer.Write(packet)
	if err != nil {
		return 0, fmt.Errorf("error when sending data to peer %s", peerId)
	}

	return n, nil
}

func (p *PeersManager) ListenAndServePeersWithProtocol(address string,
	protocol string) {

	addr := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9229,
	}

	conn, err := net.ListenUDP("udp", &addr)
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	// Do something with `conn`
}
