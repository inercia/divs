package divs

import (
	"container/list"
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/inercia/divs/server/rendezvous"
)

// A rendezvous service: something we can use for joining a groups of nodes that
// also provide a service, or for announcing ourselves...
type RendezvousService interface {
	// Announce the service, publishing with the external address provided
	Announce(string) error

	// Discover other nodes for the same service
	Discover(chan string) error

	// Leave the rendezvous
	Leave() error
}

// the peers manager is responsible for
// - establishing and keeping connections to peers
// - sending/receiving data to/from peers
type NodesManager struct {
	config *Config

	newNodesChan chan string // we'll send to this channel when we discover a new peer

	external     *External
	ExternalAddr string

	devManager   *DevManager
	rendServices *list.List
	members      *memberlist.Memberlist
}

// Create a new peers manager
func NewNodesManager(config *Config, dm *DevManager) (*NodesManager, error) {

	external, err := NewExternal(config.Global.Host, config.Global.Port)
	if err != nil {
		return nil, err
	}

	// create a list of rendezvous services
	r := list.New()
	r.PushBack(rendezvous.NewDhtService(config.Global.Serial))
	r.PushBack(rendezvous.NewMdnsService(config.Global.Serial))

	membersConfig := memberlist.DefaultWANConfig()
	members, err := memberlist.Create(membersConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create memberlist: " + err.Error())
	}

	log.Info("Creating new peers manager")
	d := NodesManager{
		config:       config,
		newNodesChan: make(chan string),
		external:     external,
		ExternalAddr: "",
		devManager:   dm,
		rendServices: r,
		members:      members,
	}

	membersConfig.Delegate = &d

	return &d, nil
}

// Start the peers manager
func (this *NodesManager) Start() (err error) {
	// calculate the external IP and port other peers will use for connecting to us
	if this.ExternalAddr, err = this.external.Obtain(); err != nil {
		return err
	}

	discovered := make(chan string)
	go func() {
		for address := range discovered {
			// Join an existing cluster by specifying at least one known member.
			n, err := this.members.Join([]string{address})
			if err != nil {
				log.Error("Failed to join cluster: " + err.Error())
			}

			this.newNodesChan <- address
		}
	}()

	for e := this.rendServices.Front(); e != nil; e = e.Next() {
		serv := e.Value.(RendezvousService)
		serv.Announce(this.ExternalAddr)
		serv.Discover(discovered)
	}

	return nil
}

// wait some time for some peers
func (d *NodesManager) Stop() (err error) {
	log.Debug("Signaling stop for discovery")
	// TODO
	return nil
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
	res := make([]byte, 0)
	return res
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed.
func (p *NodesManager) NotifyMsg([]byte) {

}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit.
func (p *NodesManager) GetBroadcasts(overhead, limit int) [][]byte {
	res := make([]byte, 0)
	return res
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (p *NodesManager) LocalState(join bool) []byte {
	res := make([]byte, 0)
	return res
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (p *NodesManager) MergeRemoteState(buf []byte, join bool) {

}
