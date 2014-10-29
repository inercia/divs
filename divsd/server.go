package divsd

import (
	"fmt"
	"sync"
)

// The DiVS server starts the nodes manager (for the p2p network), the devices
// manager (for the TAP device) and the raftd server (a combination of the
// Raft server and an HTTP server which acts as the transport)
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
		return nil, err
	}

	// Initialize the peers manager
	peersManager, err := NewNodesManager(config, devManager)
	if err != nil {
		return nil, err
	}

	// Initialize and start Raft server.
	raftServer, err := NewRaftServer(config)
	if err != nil {
		return nil, err
	}

	s = &Server{
		config:       config,
		devManager:   devManager,
		nodesManager: peersManager,
		raftServer:   raftServer,
	}

	return s, nil
}

// Starts the server.
func (s *Server) ListenAndServe() error {
	// obtain a externally-reachable IP/port for memberlist management
	defaultExternalAddr := fmt.Sprintf("%s:%d", s.config.Global.Host, s.config.Global.Port)
	membersExternalAddr, err := NewExternalUDPAddr(defaultExternalAddr)

	// start the peers manager
	if err = s.nodesManager.Start(membersExternalAddr); err != nil {
		log.Fatalf("Error when initialing peers manager: %s", err)
	}

	// and the devices manager
	if err = s.devManager.Start(); err != nil {
		log.Fatalf("Error when initialing tun/tap device manager: %s", err)
	}

	log.Debug("Are we the leader? %t", s.config.Raft.IsLeader)
	if s.config.Raft.IsLeader {
		log.Debug("... yes: we will not try to connect to anyone")
		if err := s.raftServer.InitCluster(membersExternalAddr.String()); err != nil {
			log.Fatalf("Could not initialize cluster: %v", err)
		}
		if err := s.nodesManager.WaitForNodes(); err != nil {
			return err
		}
	} else {
		log.Debug("... no: we will try to connect to the leader")
		s.nodesManager.WaitForNodes()

		// Join to leader if specified.
		log.Info("Attempting to join leader:", s.config.Raft.Leader)

		// Join an existing cluster by specifying at least one known member.
		if err := s.nodesManager.Join([]string{s.config.Raft.Leader}); err != nil {
			return err
		}
		if err := s.raftServer.JoinLeader(membersExternalAddr.String(), s.config.Raft.Leader); err != nil {
			return err
		}
	}

	log.Debug("Initializing Raft transport")
	if err := s.raftServer.StartTransport(); err != nil {
		log.Fatalf("Could not start transport: %v", err)
	}

	return nil
}
