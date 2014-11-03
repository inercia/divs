package divsd

import (
	"fmt"
	"net"
	"os"
	"sync"

	"code.google.com/p/gopacket"
	"code.google.com/p/gopacket/layers"

	"github.com/inercia/water/tuntap"
)

// the buffer length used for reading a packet from the TAP device
const TAP_BUFFER_LEN = 9000

type DevManager struct {
	numWorkers       int
	tun             *tuntap.TunTap
	nodesManager    *NodesManager
	packetsChan      chan []byte
	wg              *sync.WaitGroup
	mutex            sync.RWMutex
}

// Get all the local addresses
func GetLocalAddresses() {
	list, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for i, iface := range list {
		log.Debug("%d name=%s %v", i, iface.Name, iface)
		addrs, err := iface.Addrs()
		if err != nil {
			panic(err)
		}
		for j, addr := range addrs {
			fmt.Printf(" %d %v\n", j, addr)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////

// Create a new devices manager, in practice a TAP device manager
// This manager will be responsible for reading from the device and sending
// data to the right peers
func NewDevManager(config *Config) (d *DevManager, err error) {
	d = &DevManager{
		numWorkers    : config.Tun.NumReaders,
		packetsChan   : make(chan []byte),
		wg            : new(sync.WaitGroup),
	}
	return d, nil
}

// Set the nodes manager
func (dman *DevManager) SetNodesManager(nm *NodesManager) error {
	dman.nodesManager = nm
	return nil
}

	// Start the TAP device and start reading from it
func (dman *DevManager) Start() (err error) {
	log.Info("Initializing tap device...\n")

	euid := os.Geteuid()
	if euid != 0 {
		log.Info("WARNING: effective uid (%d) is not root: you may have not enough privileges...\n", euid)
	}

	dman.tun, err = tuntap.NewTAP("")
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	log.Info("... tap device: %s\n", dman.tun.Name())

	// Adding routines to workgroup and running then
	for i := 0; i < dman.numWorkers; i++ {
		dman.wg.Add(1)
		go dman.packetProcessor()
	}

	go dman.devReader()
	return nil
}

func (dman *DevManager) Stop() {
	// Closing channel (waiting in goroutines won't continue any more)
	close(dman.packetsChan)

	// Waiting for all goroutines to finish (otherwise they die as main routine dies)
	dman.wg.Wait()
}

// the device reader
func (dman *DevManager) devReader() {
	// Processing all packets by spreading them to `free` goroutines
	log.Debug("Starting reading from TAP device...")
	for {
		// TODO: use a sync.Pool for the buffers, so we do not generate so much garbage...

		packet := make([]byte, TAP_BUFFER_LEN)
		_, err := dman.tun.Read(packet)
		if err != nil {
			log.Info("Error reading from TAP device: %s", err)
			break
		} else {
			log.Debug("New packet read from TAP device")
			dman.packetsChan <- packet // handoff the packet to a packets processor
		}
	}
}

// Process a packet: parse it, see where it goes, etc...
func (dman *DevManager) packetProcessor() {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer dman.wg.Done()

	for packet := range dman.packetsChan {
		log.Info("Read %d", len(packet))
		packet := gopacket.NewPacket(packet, layers.LayerTypeEthernet, gopacket.Default)

		// Get the Ethernet layer from this packet
		if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
			eth, _ := ethLayer.(*layers.Ethernet)
			log.Debug("Ethernet: src:%s, dst:%s\n", eth.SrcMAC, eth.DstMAC)

			// TODO: we should parse the packet and do interesting things like
			//       - answer ARP requests
			//       - do some IGMP snooping...

			// pass the parsed packet to the nodes manager so it send it to the right destination
			dman.nodesManager.SendPacket(&EthernetPacket{*eth})
		}
	}
}
