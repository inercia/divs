package divsd

import (
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/memberlist"
	rendezvous "github.com/inercia/divs/divsd/rendezvous"
	"errors"
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

// the peers manager is responsible for
// - establishing and keeping connections to peers
// - sending/receiving data to/from peers
type NodesManager struct {
	newNodesChan      chan string // we send to this channel new, joined peers

	config            *Config
	devManager        *DevManager
	members           *memberlist.Memberlist
}

// Create a new peers manager
func NewNodesManager(config *Config, dm *DevManager) (*NodesManager, error) {
	log.Debug("Creating new nodes manager")
	d := NodesManager{
		config:        config,
		newNodesChan:  make(chan string),
		devManager:    dm,
	}
	return &d, nil
}

// Start the nodes manager
func (nm *NodesManager) Start(membersExternalAddr net.UDPAddr) (err error) {
	log.Debug("Starting nodes manager")
	membersConfig := memberlist.DefaultWANConfig()

	extIp := membersExternalAddr.IP.String()
	extPort := membersExternalAddr.Port
	log.Debug("Memberlist external IP/Port: %s:%d", extIp, extPort)
	membersConfig.BindAddr = extIp
	membersConfig.BindPort = extPort
	membersConfig.Delegate = nm
	membersConfig.LogOutput = loggerWritter
	membersConfig.EnableNotifyUnsupMsgs = true

	members, err := memberlist.Create(membersConfig)
	if err != nil {
		return fmt.Errorf("Failed to create memberlist: " + err.Error())
	}
	nm.members = members

	discovered := make(chan string)
	go func() {
		for address := range discovered {
			// Join an existing cluster by specifying at least one known member.
			numContacted, err := nm.members.Join([]string{address})
			if err != nil {
				log.Error("Failed to join node at %s: %s", address, err.Error())
			} else if numContacted > 0 {
				log.Debug("Joined %s", address)
				nm.newNodesChan <- address
			} else {
				log.Debug("No nodes contacted... but no error!!!")
			}
		}
	}()

	serviceId := nm.config.Global.Serial.ToHex()

	// create the MDNS service
	go func() {
		mdnsService, err := rendezvous.NewMdnsService("", serviceId)
		if err != nil {
			log.Error("Could not start the mDNS service")
		} else {
			mdnsService.AnnounceAndDiscover(membersExternalAddr.String(), discovered)
		}
	}()

	// create the DHT service by previously obtaining an external TCP address
	go func() {
		defaultAddr := fmt.Sprintf("0.0.0.0:%d", nm.config.Discover.Port)
		dhtAddr, err := NewExternalTCPAddr(defaultAddr)
		if err != nil {
			log.Error("Could not obtain an external port for the DHT service")
		} else {
			dhtService, err := rendezvous.NewDhtService(dhtAddr.String(), serviceId)
			if err != nil {
				log.Error("Could not start the DHT service")
			} else {
				dhtService.AnnounceAndDiscover(membersExternalAddr.String(), discovered)
			}
		}
	}()

	return nil
}

// wait some time for some peers
func (d *NodesManager) Stop() (err error) {
	log.Debug("Signaling stop for discovery")
	// TODO
	return nil
}

func (nm *NodesManager) Join(nodes []string) error {
	// Join an existing cluster by specifying at least one known member.
	n, err := nm.members.Join(nodes)
	if err != nil {
		return errors.New("Failed to join cluster: " + err.Error())
	} else {
		log.Info("Joined %d peers", n)
		return nil
	}
}

// wait some time for some peers
func (this *NodesManager) WaitForNodesTime(seconds time.Duration) (err error) {
	if this.members.NumMembers() > 0 {
		return nil
	} else {
		log.Info("Waiting for %d seconds for peers to join...", seconds)
		select {
		case <-this.newNodesChan:
			err = nil
			break
		case <-time.After(seconds * time.Second):
			err = fmt.Errorf("no valid peers found in %d seconds", seconds)
		}
		return err
	}
}

// Wait for some peers
func (this *NodesManager) WaitForNodes() error {
	if this.members.NumMembers() > 0 {
		return nil
	} else {
		log.Debug("Waiting forever for new peers to join...")
		<-this.newNodesChan
		return nil
	}
}

// Sends some data to some other node
func (this *NodesManager) SendTo(packet []byte, node *Node) error {
	log.Debug("Sending packet to %v", node)
	addr := net.IPAddr{node.Addr, ""}
	err := this.members.SendTo(&addr, packet)
	if err != nil {
		return fmt.Errorf("error when sending data to peer %s", node)
	}
	return nil
}

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (p *NodesManager) NodeMeta(limit int) []byte {
	log.Debug("Current node meta-data requested")
	res := make([]byte, 0)
	return res
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed.
func (p *NodesManager) NotifyMsg([]byte) {
	// TODO: send the message to the TAP device... maybe we should enqueue it
	// TODO: and then a worker could perform the real delivery
	log.Debug("User data received")
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit.
func (p *NodesManager) GetBroadcasts(overhead, limit int) [][]byte {
	log.Debug("Use data can be broadcasted: collecting!")
	res := make([][]byte, 0)
	return res
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (p *NodesManager) LocalState(join bool) []byte {
	log.Debug("Gathering local state for TCP Push/Pull")
	res := make([]byte, 0)
	return res
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (p *NodesManager) MergeRemoteState(buf []byte, join bool) {
	log.Debug("Merging remote state")

}
