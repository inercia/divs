package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/goraft/raft"
	"github.com/gorilla/mux"
	"github.com/inercia/divs/model/command"
	"github.com/inercia/divs/model/db"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// The raftd server is a combination of the Raft server and an HTTP
// server which acts as the transport.
type Server struct {
	name   string
	config *Config

	bindAddr     string
	externalAddr string

	peersManager *PeersManager
	devManager   *DevManager
	raftServer   raft.Server

	router     *mux.Router
	httpServer *http.Server
	db         *db.DB
	mutex      sync.RWMutex
}

// Creates a new server.
func New(config *Config) (s *Server, err error) {
	var devManager *DevManager
	var peersManager *PeersManager

	// Initialize the device manager
	devManager, err = NewDevManager(config)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the peers manager
	peersManager, err = NewPeersManager(config, devManager)
	if err != nil {
		log.Fatal(err)
	}

	s = &Server{
		config:       config,
		devManager:   devManager,
		peersManager: peersManager,
		db:           db.New(),
		router:       mux.NewRouter(),
		bindAddr:     fmt.Sprintf("http://0.0.0.0:%d", config.Global.Port),
		externalAddr: fmt.Sprintf("http://%s:%d", config.Global.Host, config.Global.Port),
	}

	s.setDataDirectory()

	// Setup commands.
	raft.RegisterCommand(&command.WriteCommand{})

	return s, nil
}

func (s *Server) setDataDirectory() {
	log.Info("Initailizaing data directory")
	if len(s.config.Raft.DataPath) == 0 {
		log.Fatalf("No data directory provided")
	}

	// Set the data directory.
	if err := os.MkdirAll(s.config.Raft.DataPath, 0744); err != nil {
		log.Critical("Unable to create path: %v", err)
	}

	log.Debug("Data directory: %s\n", s.config.Raft.DataPath)

	// Read existing name or generate a new one.
	if b, err := ioutil.ReadFile(filepath.Join(s.config.Raft.DataPath, "name")); err == nil {
		s.name = string(b)
	} else {
		s.name = fmt.Sprintf("%07x", rand.Int())[0:7]
		if err = ioutil.WriteFile(filepath.Join(s.config.Raft.DataPath, "name"),
			[]byte(s.name), 0644); err != nil {
			panic(err)
		}
	}
}

// Starts the server.
func (s *Server) ListenAndServe() error {
	var err error
	//var ips []*net.IP

	// start the peers manager
	if err = s.peersManager.Start(); err != nil {
		log.Fatalf("Error when initialing peers manager: %s", err)
	}
	log.Debug("Is leader? %t", s.config.Raft.IsLeader)
	if s.config.Raft.IsLeader {
		log.Debug("... yes: we will try to connect to the leader")
		s.peersManager.WaitForPeers()
	} else {
		log.Debug("... no: we will not try to connect to anyone")
	}

	// and the devices manager
	if err = s.devManager.Start(); err != nil {
		log.Fatalf("Error when initialing tun/tap device manager: %s", err)
	}

	// Initialize and start Raft server.
	log.Info("Initializing Raft Server: %s", s.config.Raft.DataPath)
	transporter := raft.NewHTTPTransporter("/raft", 200*time.Millisecond)
	s.raftServer, err = raft.NewServer(s.name, s.config.Raft.DataPath, transporter, nil, s.db, "")
	if err != nil {
		log.Fatal(err)
	}
	transporter.Install(s.raftServer, s)
	s.raftServer.Start()

	if s.config.Raft.IsLeader {
		// Join to leader if specified.

		log.Info("Attempting to join leader:", s.config.Raft.Leader)

		if !s.raftServer.IsLogEmpty() {
			log.Fatalf("Cannot join with an existing log")
		}
		if err := s.Join(s.config.Raft.Leader); err != nil {
			log.Fatal(err)
		}

	} else if s.raftServer.IsLogEmpty() {
		// Initialize the server by joining itself.

		log.Info("Initializing new cluster")

		_, err := s.raftServer.Do(&raft.DefaultJoinCommand{
			Name:             s.raftServer.Name(),
			ConnectionString: s.externalAddr,
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Info("Recovered from log")
	}

	log.Info("Initializing HTTP server")

	// Initialize and start HTTP server.
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Global.Port),
		Handler: s.router,
	}

	s.router.HandleFunc("/db/{key}", s.readHandler).Methods("GET")
	s.router.HandleFunc("/db/{key}", s.writeHandler).Methods("POST")
	s.router.HandleFunc("/join", s.joinHandler).Methods("POST")

	log.Info("Listening at:", s.bindAddr)

	return s.httpServer.ListenAndServe()
}

// This is a hack around Gorilla mux not providing the correct net/http
// HandleFunc() interface.
func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.router.HandleFunc(pattern, handler)
}

/////////////////////////////////////////////////////////////////////////////////

// Joins to the leader of an existing cluster.
func (s *Server) Join(leader string) error {
	command := &raft.DefaultJoinCommand{
		Name:             s.raftServer.Name(),
		ConnectionString: s.externalAddr,
	}

	var b bytes.Buffer
	json.NewEncoder(&b).Encode(command)
	resp, err := http.Post(fmt.Sprintf("http://%s/join", leader), "application/json", &b)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func (s *Server) joinHandler(w http.ResponseWriter, req *http.Request) {
	command := &raft.DefaultJoinCommand{}

	if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := s.raftServer.Do(command); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) readHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	value := s.db.Get(vars["key"])
	w.Write([]byte(value))
}

func (s *Server) writeHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	// Read the value from the POST body.
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	value := string(b)

	// Execute the command against the Raft server.
	_, err = s.raftServer.Do(command.NewWriteCommand(vars["key"], value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
