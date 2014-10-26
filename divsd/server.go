package divsd

import (
	"fmt"
	"github.com/hashicorp/memberlist"
	"sync"
)

// The raftd server is a combination of the Raft server and an HTTP
// server which acts as the transport.
type Server struct {
	config *Config

	peersManager *NodesManager
	devManager   *DevManager
	memberlist   *memberlist.Memberlist
	raftServer   *RaftServer

	mutex sync.RWMutex
}

// Creates a new server.
func New(config *Config) (s *Server, err error) {
	var (
		devManager   *DevManager
		peersManager *NodesManager
		mlist        *memberlist.Memberlist
	)

	//bindAddr := fmt.Sprintf("http://0.0.0.0:%d", config.Global.Port)
	externalAddr := fmt.Sprintf("http://%s:%d", config.Global.Host, config.Global.Port)

	// Initialize the device manager
	devManager, err = NewDevManager(config)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the peers manager
	peersManager, err = NewNodesManager(config, devManager)
	if err != nil {
		log.Fatal(err)
	}

	/* Create the initial memberlist from a safe configuration.
	   Please reference the godoc for other default config types.
	   http://godoc.org/github.com/hashicorp/memberlist#Config
	*/
	log.Info("Creating memberlist config")
	mlist, err = memberlist.Create(memberlist.DefaultLocalConfig())
	if err != nil {
		log.Fatal("Failed to create memberlist: " + err.Error())
	}

	// Initialize and start Raft server.
	log.Info("Initializing Raft Server: %s", s.config.Raft.DataPath)
	s.raftServer, err = NewRaftServer(config, externalAddr)
	if err != nil {
		log.Fatal(err)
	}

	s = &Server{
		config:       config,
		devManager:   devManager,
		peersManager: peersManager,
		memberlist:   mlist,
	}

	return s, nil
}

// Starts the server.
func (s *Server) ListenAndServe() error {
	var err error

	// start the peers manager
	if err = s.peersManager.Start(); err != nil {
		log.Fatalf("Error when initialing peers manager: %s", err)
	}
	log.Debug("Is leader? %t", s.config.Raft.IsLeader)
	if s.config.Raft.IsLeader {
		log.Debug("... yes: we will try to connect to the leader")
		s.peersManager.WaitForNodes()
	} else {
		log.Debug("... no: we will not try to connect to anyone")
	}

	// and the devices manager
	if err = s.devManager.Start(); err != nil {
		log.Fatalf("Error when initialing tun/tap device manager: %s", err)
	}

	if !s.config.Raft.IsLeader {
		// Join to leader if specified.
		log.Info("Attempting to join leader:", s.config.Raft.Leader)

		// Join an existing cluster by specifying at least one known member.
		n, err := s.memberlist.Join([]string{s.config.Raft.Leader})
		if err != nil {
			log.Fatalf("Failed to join cluster: " + err.Error())
		} else {
			log.Info("Joined %d peers", n)
		}

		if err := s.raftServer.JoinLeader(s.config.Raft.Leader); err != nil {
			log.Fatalf("Could not join %s: %v", s.config.Raft.Leader, err)
		}
	} else if err := s.raftServer.InitCluster(); err != nil {
		log.Fatalf("Could not initialize cluster: %v", err)
	}

	log.Debug("Initializing Raft transport")
	if err := s.raftServer.StartTransport(); err != nil {
		log.Fatalf("Could not start transport: %v", err)
	}

	return nil
}
