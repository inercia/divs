package divs

import (
	"fmt"
	"net"
	"strconv"

	"github.com/hashicorp/memberlist"
)

type RawPacket *[]byte

const DEFAULT_PROTOCOL = "udp"

const DEFAULT_POOL_MIN = 5

const DEFAULT_POOL_MAX = 10

// maybe we should use this in the future:
// http://zhen.org/blog/ring-buffer-variable-length-low-latency-disruptor-style/

// maybe we should send with ratelimit
// https://github.com/juju/ratelimit

type Node struct {
	memberlist.Node

	recvChan chan []byte
	sendChan chan []byte
}

// @param address: a string in the form IP:port
func NewNode(address string) (*Node, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("could not parse peer address: %v", err)
	}

	iPort, pErr := strconv.Atoi(port)
	if pErr != nil {
		return nil, fmt.Errorf("could not parse port number from address")
	}

	// TODO

	p := &Node{
		sendChan: make(chan []byte),
		recvChan: make(chan []byte),
	}

	// create the two workers that will receive/send data
	go p.sendWorker(p.sendChan)
	go p.receiveWorker(p.recvChan)

	return p, nil
}

/////////////////////////////////////////////////////////////////////////////
// read/write/close interface
/////////////////////////////////////////////////////////////////////////////

func (p *Node) Read(data []byte) (n int, err error) {
	// TODO
	return 0, nil
}

func (p *Node) Write(data []byte) (n int, err error) {
	p.sendChan <- data
	return len(data), nil
}

func (p *Node) Close() error {
	// close pool any time you want, this closes all the connections inside a pool
	return nil
}

/////////////////////////////////////////////////////////////////////////////
// send & receive workers
/////////////////////////////////////////////////////////////////////////////

// start a goroutine that reads from a peer
func (p *Node) sendWorker(sendChan chan []byte) {
	// now you can get a connection from the pool, if there is no connection
	// available it will create a new one via the factory function.
	log.Info("Starting sender worker for %s")
	for data := range sendChan {
		n, err := conn.Write(data)
		if err != nil {
			log.Info("ERROR: when sending data to %s", p.Id)
		} else {
			log.Info("%d bytes sent to %s", n, p.Id)
		}
	}

	// do something with conn and put it back to the pool by closing the connection
	// (this doesn't close the underlying connection instead it's putting it back
	// to the pool).
	conn.Close()
}

// start a goroutine that reads from a peer
func (p *Node) receiveWorker(sendChan chan []byte) {
	log.Info("Starting receiver worker for %s", p.Id)
	// TODO
}
