package divsd

import (
	"errors"
	"fmt"
	"net"
	"time"
	"github.com/hashicorp/memberlist"
	"github.com/inercia/divs/divsd/rendezvous"
)

// joined channel length
const JOINED_CHAN_LEN = 10

// discovered channel length
const DISCOVERED_CHAN_LEN = 10

// Unknown destination mac
var ERR_UNKNOWN_DST_MAC = fmt.Errorf("Unknown destination mac")

// the peers manager is responsible for
// - establishing and keeping connections to peers
// - sending/receiving data to/from peers
type NodesManager struct {
	config         *Config
	devManager     *DevManager
	members        *memberlist.Memberlist
	membersExtAddr net.UDPAddr

	discoveredChan chan string // we send to this channel possible, discovered peers
	joinedChan     chan string // we send to this channel new, joined peers

	macsToNodes map[string]*Node
}

// Create a new peers manager
func NewNodesManager(config *Config) (*NodesManager, error) {
	log.Debug("Creating new nodes manager")
	d := NodesManager{
		config:         config,
		joinedChan:     make(chan string, JOINED_CHAN_LEN),
		discoveredChan: make(chan string, DISCOVERED_CHAN_LEN),
		macsToNodes:    make(map[string]*Node),
	}
	return &d, nil
}

// Set the devices manager
func (nm *NodesManager) SetDevManager(dm *DevManager) error {
	nm.devManager = dm
	return nil
}

// Start the nodes manager
func (nm *NodesManager) Start(membersExtAddr net.UDPAddr) (err error) {
	log.Debug("Starting nodes manager")
	nm.membersExtAddr = membersExtAddr

	extIp := nm.membersExtAddr.IP.String()
	extPort := nm.membersExtAddr.Port
	log.Debug("Memberlist external IP/Port: %s:%d", extIp, extPort)

	membersConfig := memberlist.DefaultWANConfig()
	membersConfig.BindAddr = nm.config.Global.BindIP
	membersConfig.BindPort = extPort
	membersConfig.Delegate = nm
	membersConfig.Events = nm
	membersConfig.LogOutput = loggerWritter

	members, err := memberlist.Create(membersConfig)
	if err != nil {
		return fmt.Errorf("Failed to create memberlist: " + err.Error())
	}
	nm.members = members

	// start reading from the "discoveredChan" channel and, for each new peer
	// discovered, instruct the "memberlist" to "join" it
	go func() {
		for address := range nm.discoveredChan {
			go func(a string) {
				// Join an existing cluster by specifying at least one known member.
				_, err := nm.members.Join([]string{a})
				if err != nil {
					log.Error("Failed to join node at %s: %s", a, err.Error())
				}
				// we will continue in NotifyJoin()...
			}(address)
		}
	}()

	serviceId := nm.config.Global.Serial.ToHex()
	bindIp := nm.config.Global.BindIP
	dhtPort := nm.config.Discover.Port
	rendezvous.Start(serviceId, bindIp, dhtPort, nm.membersExtAddr.String(), nm.discoveredChan)

	return nil
}

// wait some time for some peers
func (nm *NodesManager) Stop() (err error) {
	log.Debug("Signaling stop for discovery")
	close(nm.discoveredChan)
	return nil
}

// Join a new peer
// This method is invoked when we have detected a new peer with the rendezvous
// subsystem. It triggers the `memberlist` join.
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

// Wait some time for some peers
func (nm *NodesManager) WaitForNodesTime(seconds time.Duration) (err error) {
	if nm.members.NumMembers() == 0 {
		log.Info("Waiting for %d seconds for peers to join...", seconds)
		select {
		case <-nm.joinedChan:
			break
		case <-time.After(seconds * time.Second):
			return ERR_TIMEOUT_PEERS
		}
	}
	return nil
}

// Wait for some peers
func (this *NodesManager) WaitForNodes() error {
	if this.members.NumMembers() == 0 {
		log.Debug("Waiting for a new peer...")
		<-this.joinedChan
	}
	return nil
}

// Wait for some peers
func (this *NodesManager) WaitForNodesForever() error {
	log.Debug("Waiting for peers to be discovered...")
	for {
		<-this.joinedChan
		log.Debug("[WaitForNodesForever] node joined")
	}
	return nil
}

// Sends a packet to the corresponding Node
// If no valid Node is found for this packet, it is silently discarded
func (nm *NodesManager) SendPacket(packet *EthernetPacket) error {
	// check if we have a valid destination node for this packet
	destMac := packet.DstMAC.String()
	node, found := nm.macsToNodes[destMac]
	if !found {
		log.Debug("Trying to send to unknown peer %s", node)
		return ERR_UNKNOWN_DST_MAC
	}
	return node.Send(packet)
}

// Sends some data to some other node
func (nm *NodesManager) SendTo(packet []byte, node *Node) error {
	log.Debug("Sending packet to %v", node)
	addr := net.IPAddr{node.Addr, ""}
	err := nm.members.SendTo(&addr, packet)
	if err != nil {
		return fmt.Errorf("error when sending data to peer %s", node)
	}
	return nil
}

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (nm *NodesManager) NodeMeta(limit int) []byte {
	log.Debug("Current node meta-data requested")
	res := make([]byte, 0)
	return res
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed.
func (nm *NodesManager) NotifyMsg(buf []byte) {
	log.Debug("User data received")
	messageType, message, err := getTypeAndEncodedMsg(buf)
	if err != nil {
		log.Error("Could not receive message: %s", err)
		return
	}

	switch messageType {
	case MSG_DIVS_PKG_ETH:
		log.Debug("Data packet received: %d bytes", len(message))
		// TODO: send the message to the TAP device... maybe we should enqueue it
		// TODO: and then a worker could perform the real delivery
	default:
		log.Error("Unknown message received: %s", messageType)
	}
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit.
func (nm *NodesManager) GetBroadcasts(overhead, limit int) [][]byte {
	log.Debug("Use data can be broadcasted: collecting!")
	res := make([][]byte, 0)
	return res
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (nm *NodesManager) LocalState(join bool) []byte {
	res := make([]byte, 0)
	if join {
		log.Debug("Gathering local state for joining")
		// TODO: local state info
	} else {
		log.Debug("Gathering local state for TCP Push/Pull")
		// TODO: local state info
	}
	return res
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (nm *NodesManager) MergeRemoteState(buf []byte, join bool) {
	log.Debug("[MergeRemoteState] merging remote state")
	// TODO: merge remote state
}

// NotifyJoin is invoked when a node is detected to have joined the memberlist.
// The Node argument must not be modified.
func (nm *NodesManager) NotifyJoin(node *memberlist.Node) {
	newNodeAddr := fmt.Sprintf("%s:%d", node.Addr, node.Port)
	log.Debug("[NotifyJoin] new node joined: %s", newNodeAddr)
	nm.joinedChan <- newNodeAddr
	// TODO: something else to do when someone else joins?
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (nm *NodesManager) NotifyLeave(node *memberlist.Node) {
	log.Debug("[NotifyLeave] node %s has been declared as unreachable", node)
	// TODO: remove all the MACs for this node that has left
	for savedMac, savedNode := range nm.macsToNodes {
		if savedNode.Equal(node) {
			delete(nm.macsToNodes, savedMac)
		}
	}
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (nm *NodesManager) NotifyUpdate(node *memberlist.Node) {
	log.Debug("[NotifyUpdate] node %s has updated", node)
	// TODO: what should we do here?
}
