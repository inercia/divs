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

	var log = logging.MustGetLogger("divs")

	options := struct {
			ConfigPath string `goptions:"-c, --config, config, description='Config file name'"`

			Host       string `goptions:"--host, maps='Global/Host', description='External hostname/IP to announce to peers'"`
			Port       int    `goptions:"--port, maps='Global/Port', description='Force the external port for peers connecting'"`

		// raft
			RaftLeader string `goptions:"--join, maps='Raft/Leader', description='host:port of leader to join'"`
			DataPath   string `goptions:"--data, maps='Raft/DataPath', description='data path directory'"`

		// discovery
			DiscoverPort bool   `goptions: "--dport, maps='Discover/Port', description='discovery protocol port'"`
			Create       bool   `goptions: "--create, description='create a new switch'"`
			Join         string `goptions: "--serial, maps='Global/Serial',
		                     description='switch serial number to join'"`

		// other
			Timeout time.Duration `goptions:"-t, --timeout, description='connection timeout in seconds'"`
			Pidfile string        `goptions:"--pidfile, description='PID file'"`

		// aux
			Help    goptions.Help `goptions:"-h, --help, description='show this help'"`
			Verbose bool          `goptions:"-v, --verbose"`
			Trace   bool          `goptions:"--trace"`
			Debug   bool          `goptions:"--debug"`
		}{ // Default values goes here
		Create:   true,
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

	// check if we are creating a new switch or just joining an existing one.
	// if we are creating one, we must create a new serial...
	if config.Global.Serial.Empty() {
		if options.Create {
			config.Global.Serial = divsd.NewSwitchId()
			config.Raft.IsLeader = true
		} else {
			log.Critical("# Error: would try to join a switch but no serial provided")
			os.Exit(1)
		}
	} else {
		if options.Create {
			log.Critical("# Error: cannot provide serial when creating switch")
			os.Exit(1)
		} else {
			config.Raft.IsLeader = false
		}
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

	rand.Seed(time.Now().UnixNano())

	s, err := divsd.New(config)
	log.Fatal(s.ListenAndServe())
}
