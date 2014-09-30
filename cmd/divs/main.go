package main

import (
	"github.com/goraft/raft"
	"github.com/inercia/divs/command"
	"github.com/inercia/divs/server"
	"github.com/voxelbrain/goptions"
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	log.SetFlags(0)

	options := struct {
		Host string `goptions:"-h, --host, description='Hostname'"`
		Port int    `goptions:"-p, --port, description='Port'"`
		Join string `goptions:"-j, --join, obligatory, description='host:port of leader to join'"`
		Path string `goptions:"-d, --data, obligatory, description='data path directory'"`

		// other
		Timeout time.Duration `goptions:"-t, --timeout, description='Connection timeout in seconds'"`
		Help    goptions.Help `goptions:"-h, --help, description='Show this help'"`
		Verbose bool          `goptions:"-v, --verbose"`
		Trace   bool          `goptions:"-t, --trace"`
		Debug   bool          `goptions:"-d, --debug"`
	}{ // Default values goes here
		Timeout: 10 * time.Second,
		Port:    4004,
	}

	goptions.ParseAndFail(&options)

	if options.Verbose {
		log.Print("Verbose logging enabled.")
	}
	if options.Trace {
		raft.SetLogLevel(raft.Trace)
		log.Print("Raft trace debugging enabled.")
	} else if options.Debug {
		raft.SetLogLevel(raft.Debug)
		log.Print("Raft debugging enabled.")
	}

	rand.Seed(time.Now().UnixNano())

	// Setup commands.
	raft.RegisterCommand(&command.WriteCommand{})

	// Set the data directory.
	if err := os.MkdirAll(options.Path, 0744); err != nil {
		log.Fatalf("Unable to create path: %v", err)
	}

	log.SetFlags(log.LstdFlags)
	s := server.New(options.Path, options.Host, options.Port)
	log.Fatal(s.ListenAndServe(options.Join))
}
