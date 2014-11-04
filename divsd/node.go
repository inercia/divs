package divsd

import (
	"fmt"
	"net"
	"strconv"

	"bytes"
	"github.com/hashicorp/memberlist"
)

// maybe we should use this in the future:
// http://zhen.org/blog/ring-buffer-variable-length-low-latency-disruptor-style/

// maybe we should send with ratelimit
// https://github.com/juju/ratelimit

// the send queue length used for sending to a node
const SEND_QUEUE_LEN = 100

type Node struct {
	*memberlist.Node

	manager  *NodesManager
	sendChan chan Encodeable
}

// @param address: a string in the form IP:port
func NewNode(address string, nm *NodesManager) (*Node, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("could not parse peer address: %v", err)
	}
	portI, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("could not parse port number from address")
	}

	n := Node{
		Node: &memberlist.Node{
			Addr: net.ParseIP(host),
			Port: uint16(portI),
		},
		sendChan: make(chan Encodeable, SEND_QUEUE_LEN),
		manager:  nm,
	}

	// create a worker for sending data
	go n.sendWorker()

	return &n, nil
}

// Send some serializable object to this node
// Data is enqueued in a queue for sending
// This method will only be invoked from the NodesManager
func (node *Node) Send(data Encodeable) error {
	log.Debug("Enqueuing data for sending to %v", node)
	node.sendChan <- data
	return nil
}

// Close the node, releasing all the resources associated with it
func (node *Node) Close() error {
	log.Debug("Closing node %s", node)
	close(node.sendChan)
	return nil
}

// Start a coroutine that send to this node
func (node *Node) sendWorker() {
	ipAddr, err := net.ResolveIPAddr("ip", node.Addr.String())
	if err != nil {
		log.Debug("Could not parse IP address %v", node.Addr)
	}

	log.Info("Starting sender worker for %s", ipAddr)
	for data := range node.sendChan {
		marshaled, err := data.Encode()
		if err != nil {
			log.Debug("Error encoding data for %s: %s", node.Addr, err)
		}
		err = node.manager.members.SendTo(ipAddr, marshaled)
		if err != nil {
			log.Debug("Error sending to %s: %s", node.Addr, err)
		}
	}
}

// Compare to another node, returning "true" if they are equal
func (node Node) Equal(other *memberlist.Node) bool {
	if bytes.Compare(node.Node.Addr, other.Addr) != 0 {
		return false
	}
	if node.Node.Port != other.Port {
		return false
	}
	return true
}
