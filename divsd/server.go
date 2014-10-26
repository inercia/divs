package divsd

import (
	"fmt"
	"sync"
)

// The raftd server is a combination of the Raft server and an HTTP
// server which acts as the transport.
type Server struct {
	config *Config

	nodesManager *NodesManager
	devManager   *DevManager
	raftServer   *RaftServer

	mutex sync.RWMutex
}

// Creates a new server.
func New(config *Config) (s *Server, err error) {
	// Initialize the device manager
	devManager, err := NewDevManager(config)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the peers manager
	peersManager, err := NewNodesManager(config, devManager)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize and start Raft server.
	log.Info("Initializing Raft Server: %s", s.config.Raft.DataPath)
	s.raftServer, err = NewRaftServer(config)
	if err != nil {
		log.Fatal(err)
	}

	s = &Server{
		config:       config,
		devManager:   devManager,
		nodesManager: peersManager,
	}

	return s, nil
}

// Starts the server.
func (s *Server) ListenAndServe() error {
	// obtain a externally-reachable IP/port for memberlist management
	membersExternalAddr, err := NewExternalUDPAddr(fmt.Sprintf("%s:%d",
		s.config.Global.Host, s.config.Global.Port))

	// start the peers manager
	if err = s.nodesManager.Start(membersExternalAddr); err != nil {
		log.Fatalf("Error when initialing peers manager: %s", err)
	}
	log.Debug("Is leader? %t", s.config.Raft.IsLeader)
	if s.config.Raft.IsLeader {
		log.Debug("... yes: we will try to connect to the leader")
		s.nodesManager.WaitForNodes()
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
		s.nodesManager.Join([]string{s.config.Raft.Leader})

		if err := s.raftServer.JoinLeader(membersExternalAddr.String(), s.config.Raft.Leader); err != nil {
			log.Fatalf("Could not join %s: %v", s.config.Raft.Leader, err)
		}
	} else if err := s.raftServer.InitCluster(membersExternalAddr.String()); err != nil {
		log.Fatalf("Could not initialize cluster: %v", err)
	}

	log.Debug("Initializing Raft transport")
	if err := s.raftServer.StartTransport(); err != nil {
		log.Fatalf("Could not start transport: %v", err)
	}

	return nil
}
