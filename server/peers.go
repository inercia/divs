package server

import (
	"fmt"
	"gopkg.in/fatih/pool.v2"
	"net"
	"strconv"
)

type RawPacket *[]byte

const DEFAULT_PROTOCOL = "udp"

const DEFAULT_POOL_MIN = 5

const DEFAULT_POOL_MAX = 10

// maybe we should use this in the future:
// http://zhen.org/blog/ring-buffer-variable-length-low-latency-disruptor-style/

type Peer struct {
	Id       string // the id is just the IP:port
	Host     net.IP
	Port     int
	Protocol string

	connPool pool.Pool

	recvChan chan []byte
	sendChan chan []byte
}

// @param address: a string in the form IP:port
func NewPeer(address string) (*Peer, error) {
	return NewPeerWithProtocol(address, DEFAULT_PROTOCOL)
}

// @param address: a string in the form IP:port
// @param protocl: a protocol, like "tcp", "udp", etc...
func NewPeerWithProtocol(address string, protocol string) (*Peer, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("could not parse peer address: %v", err)
	}

	iPort, pErr := strconv.Atoi(port)
	if pErr != nil {
		return nil, fmt.Errorf("could not parse port number from address")
	}

	// create a factory() to be used with channel based pool
	factory := func() (net.Conn, error) {
		log.Debug("Creating %s connection to %s...", protocol, address)
		return net.Dial(protocol, address)
	}

	// create a new channel based pool with an initial capacity of 5 and maximum
	// capacity of 30. The factory will create 5 initial connections and put it
	// into the pool.
	connPool, err := pool.NewChannelPool(DEFAULT_POOL_MIN, DEFAULT_POOL_MAX, factory)
	if err != nil {
		return nil, fmt.Errorf("could not create connections pool")
	}

	p := &Peer{
		Id:       address,
		Host:     net.ParseIP(host),
		Port:     iPort,
		Protocol: protocol,
		connPool: connPool,
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

func (p *Peer) Read(data []byte) (n int, err error) {
	// TODO
	return 0, nil
}

func (p *Peer) Write(data []byte) (n int, err error) {
	p.sendChan <- data
	return len(data), nil
}

func (p *Peer) Close() error {
	// close pool any time you want, this closes all the connections inside a pool
	p.connPool.Close()
	return nil
}

/////////////////////////////////////////////////////////////////////////////
// send & receive workers
/////////////////////////////////////////////////////////////////////////////

// start a goroutine that reads from a peer
func (p *Peer) sendWorker(sendChan chan []byte) {
	// now you can get a connection from the pool, if there is no connection
	// available it will create a new one via the factory function.
	conn, getErr := p.connPool.Get()
	if getErr != nil {
		log.Info("ERROR: could not get a connection to %s", p.Id)
		return
	}

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
func (p *Peer) receiveWorker(sendChan chan []byte) {
	log.Info("Starting receiver worker for %s", p.Id)
	// TODO
}
