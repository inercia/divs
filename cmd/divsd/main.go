package main

import (
	"github.com/facebookgo/pidfile"
	"github.com/goraft/raft"
	"math/rand"
	"os"
	"time"
	logging "github.com/op/go-logging"

	"github.com/inercia/goptions"
	"github.com/inercia/divs/divsd"
)

func main() {
	// command line options
	// note: do not break lines in goptions [alvaro]
	options := struct {
			ConfigPath string `goptions:"-c, --config, config, description='config file name'"`

		// external IP/port
			Host       string `goptions:"--host, maps='Global/Host', description='forced external hostname/IP to announce to peers'"`
			Port       int    `goptions:"--port, maps='Global/Port', description='forced external port to announce to peers'"`

		// raft
			RaftLeader string `goptions:"--leader, maps='Raft/Leader', description='forced host:port of leader to join'"`
			DataPath   string `goptions:"--data, maps='Raft/DataPath', description='data path directory'"`

		// switch
			Create     bool   `goptions:"--create, description='create a new virtual switch'"`
			Serial     string `goptions:"--join, description='virtual switch serial number to join'"`

		// discovery
			DiscoverPort int    `goptions:"--dhtport, maps='Discover/Port', description='discovery protocol port'"`

		// other
			Timeout time.Duration `goptions:"-t, --timeout, description='connection timeout in seconds'"`
			Pidfile string        `goptions:"--pidfile, description='file where the PID will be saved'"`

		// aux
			Help    goptions.Help `goptions:"-h, --help, description='show this help'"`
			Verbose bool          `goptions:"-v, --verbose"`
			Trace   bool          `goptions:"--trace"`
			Debug   bool          `goptions:"--debug"`
		}{ // Default values goes here
		Create:   false,
		Serial:   "",
		Timeout:  DEFAULT_TIMEOUT,
		Host:     "",
		DataPath: DEFAULT_DATA_PATH,
		Pidfile:  "",
	}

	goptions.ParseAndFail(&options)
	config := divsd.NewConfig()
	err := goptions.LoadConf(config)
	if err != nil {
		log.Critical("# Error: when loading config file: %s", err)
		os.Exit(1)
	}

	if len(options.Pidfile) > 0 {
		pidfile.SetPidfilePath(options.Pidfile)
		if err := pidfile.Write(); err != nil {
			log.Fatal(err)
		}
	}

	if options.Verbose {
		log.Info("Verbose logging enabled.")
		logging.SetLevel(logging.DEBUG, divsd.LOG_MODULE)
	}
	if options.Trace {
		raft.SetLogLevel(raft.Trace)
		log.Info("Raft trace debugging enabled.")
	} else if options.Debug {
		raft.SetLogLevel(raft.Debug)
		log.Info("Raft debugging enabled.")
	}

	// check if we are creating a new switch or just joining an existing one.
	// if we are creating one, we must create a new serial...
	if options.Create || len(options.Serial) == 0 {
		config.Global.Serial = divsd.NewSwitchId()
		config.Raft.IsLeader = true
		log.Info("Creating virtual switch ID:%s", config.Global.Serial)
	} else {
		log.Info("Will join virtual switch %s", options.Serial)
		config.Global.Serial = divsd.NewSwitchFromString(options.Serial)
		config.Raft.IsLeader = false
	}
	rand.Seed(time.Now().UnixNano())

	s, err := divsd.New(config)
	log.Fatal(s.ListenAndServe())
}
