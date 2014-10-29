package divsd

import (
	"fmt"
	"github.com/inercia/water/tuntap"
	"net"
	"os"
	"sync"
)

type DevManager struct {
	numWorkers       int
	tun             *tuntap.TunTap
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

func (dman *DevManager) devReader() {
	// Processing all packets by spreading them to `free` goroutines
	log.Debug("Starting reading from TAP device...")
	for {
		packet := make([]byte, 1500)
		_, err := dman.tun.Read(packet)
		if err != nil {
			log.Info("Error reading from TAP device: %s", err)
			break
		} else {
			log.Debug("New packet read from TAP device")
			// TODO: split/parse... the packet and send it to the righ worker
			dman.packetsChan <- packet
		}
	}
}

// Process a packet: parse it, see where it goes, etc...
func (dman *DevManager) packetProcessor() {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer dman.wg.Done()

	for packet := range dman.packetsChan {
		// TODO: do the job here
		log.Info("Read %d", len(packet))
	}
}
