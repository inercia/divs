package divsd

import (
	"fmt"
	"github.com/inercia/water/tuntap"
	"net"
	"os"
	"sync"
)

type DevManager struct {
	numWorkers int
	tun        *tuntap.TunTap
	mutex      sync.RWMutex
}

func GetLocalAddresses() {
	list, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for i, iface := range list {
		fmt.Printf("%d name=%s %v\n", i, iface.Name, iface)
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

// create a new Tap device manager
// this manager will be responsible for reading from the device and sending
// data to the right peers
func NewDevManager(config *Config) (d *DevManager, err error) {
	d = &DevManager{
		numWorkers: config.Tun.NumReaders,
	}
	return d, nil
}

func (d *DevManager) devReader(packetsChan chan []byte, wg *sync.WaitGroup) {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer wg.Done()

	for packet := range packetsChan {
		// Do the job here
		log.Info("Read %d", len(packet))
	}
}

// initialize the tap device and start reading from it
func (d *DevManager) Start() (err error) {
	log.Info("Initializing tap device...\n")

	euid := os.Geteuid()
	if euid != 0 {
		log.Info("WARNING: effective uid (%d) is not root: you may have not enough privileges...\n", euid)
	}

	d.tun, err = tuntap.NewTAP("")
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	log.Info("... tap device: %s\n", d.tun.Name())

	pCh := make(chan []byte)
	wg := new(sync.WaitGroup)

	// Adding routines to workgroup and running then
	for i := 0; i < d.numWorkers; i++ {
		wg.Add(1)
		go d.devReader(pCh, wg)
	}

	// Processing all packets by spreading them to `free` goroutines
	for {
		packet := make([]byte, 1500)
		_, err := d.tun.Read(packet)
		if err != nil {
			log.Info("Error reading from tap device: %s", err)
			break
		}
		// TODO: split/parse... the packet and send it to the workers
		pCh <- packet
	}

	// Closing channel (waiting in goroutines won't continue any more)
	close(pCh)

	// Waiting for all goroutines to finish (otherwise they die as main routine dies)
	wg.Wait()

	return nil
}
