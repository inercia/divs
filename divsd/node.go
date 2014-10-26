package divsd

import (
	"fmt"
	"net"
	"strconv"

	"github.com/hashicorp/memberlist"
)

// maybe we should use this in the future:
// http://zhen.org/blog/ring-buffer-variable-length-low-latency-disruptor-style/

// maybe we should send with ratelimit
// https://github.com/juju/ratelimit

type Node struct {
	*memberlist.Node
	*net.UDPAddr

	memberlist *memberlist.Memberlist

	recvChan chan []byte
	sendChan chan []byte
}

// @param address: a string in the form IP:port
func NewNode(address string, ml *memberlist.Memberlist) (*Node, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("could not parse peer address: %v", err)
	}
	portI, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("could not parse port number from address")
	}
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, fmt.Errorf("could not parse address %s", address)
	}

	// TODO

	p := Node{
		Node: &memberlist.Node{
			Addr: net.ParseIP(host),
			Port: uint16(portI),
		},
		UDPAddr : udpAddr,
		sendChan: make(chan []byte),
		recvChan: make(chan []byte),
		memberlist: ml,
	}

	// create the two workers that will receive/send data
	go p.sendWorker(p.sendChan)

	return &p, nil
}

/////////////////////////////////////////////////////////////////////////////
// send worker
/////////////////////////////////////////////////////////////////////////////

// start a goroutine that reads from a peer
func (p *Node) sendWorker(sendChan chan []byte) {
	// now you can get a connection from the pool, if there is no connection
	// available it will create a new one via the factory function.
	log.Info("Starting sender worker for %s")
	for {
		data, available := <-sendChan
		if !available {
			break
		}
		if err := p.memberlist.SendTo(p, data); err != nil {
			log.Error("ERROR: when sending data to %s", p.Name)
		} else {
			log.Debug("Sent %d bytes to %s", len(data), p.Name)
		}
	}
}
