package divs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goraft/raft"
	"github.com/gorilla/mux"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/inercia/divs/server/raft/command"
	"github.com/inercia/divs/server/raft/db"
)

var (
	ERR_LOG_NOT_EMPTY = errors.New("log not empty")
)

type RaftServer struct {
	name         string
	config       *Config
	externalAddr string

	raftServer raft.Server
	router     *mux.Router
	httpServer *http.Server
	db         *db.DB
}

// Create a new Raft server
func NewRaftServer(config *Config, externalAddr string) (*RaftServer, error) {
	s := RaftServer{
		config:       config,
		db:           db.New(),
		router:       mux.NewRouter(),
		externalAddr: externalAddr,
	}

	log.Info("Initailizaing data directory")
	if len(config.Raft.DataPath) == 0 {
		log.Fatalf("No data directory provided")
	}

	// Set the data directory.
	if err := os.MkdirAll(config.Raft.DataPath, 0744); err != nil {
		log.Critical("Unable to create path: %v", err)
	}

	log.Debug("Data directory: %s\n", config.Raft.DataPath)

	// Read existing name or generate a new one.
	if b, err := ioutil.ReadFile(filepath.Join(config.Raft.DataPath, "name")); err == nil {
		s.name = string(b)
	} else {
		s.name = fmt.Sprintf("%07x", rand.Int())[0:7]
		if err = ioutil.WriteFile(filepath.Join(config.Raft.DataPath, "name"), []byte(s.name), 0644); err != nil {
			log.Fatal(err)
		}
	}

	var e error
	transporter := raft.NewHTTPTransporter("/raft", 200*time.Millisecond)
	if s.raftServer, e = raft.NewServer(s.name, s.config.Raft.DataPath, transporter, nil, s.db, ""); e != nil {
		log.Fatal(e)
	}

	transporter.Install(s.raftServer, &s)
	s.raftServer.Start()

	// Setup commands.
	raft.RegisterCommand(&command.WriteCommand{})

	return &s, nil
}

// Joins to the leader of an existing cluster.
func (s *RaftServer) JoinLeader(leader string) error {
	if !s.raftServer.IsLogEmpty() {
		log.Fatalf("Cannot join with an existing log")
	}

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

func (s *RaftServer) InitCluster() error {
	if !s.raftServer.IsLogEmpty() {
		return ERR_LOG_NOT_EMPTY
	}

	// Initialize the server by joining itself.
	log.Info("Initializing new cluster")
	_, err := s.raftServer.Do(&raft.DefaultJoinCommand{
		Name:             s.raftServer.Name(),
		ConnectionString: s.externalAddr,
	})
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (s *RaftServer) IsLogEmpty() bool {
	return s.raftServer.IsLogEmpty()
}

///////////////////////////////////////
// transport related functions
///////////////////////////////////////

func (s *RaftServer) StartTransport() error {
	log.Info("Initializing HTTP server")
	listenAddr := fmt.Sprintf(":%d", s.config.Global.Port)

	// Initialize and start HTTP server.
	s.httpServer = &http.Server{
		Addr:    listenAddr,
		Handler: s.router,
	}

	s.router.HandleFunc("/db/{key}", s.readHandler).Methods("GET")
	s.router.HandleFunc("/db/{key}", s.writeHandler).Methods("POST")
	s.router.HandleFunc("/join", s.joinHandler).Methods("POST")

	log.Debug("Listening at:", listenAddr)
	return s.httpServer.ListenAndServe()
}

// This is a hack around Gorilla mux not providing the correct net/http
// HandleFunc() interface.
func (s *RaftServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.router.HandleFunc(pattern, handler)
}

func (s *RaftServer) joinHandler(w http.ResponseWriter, req *http.Request) {
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

func (s *RaftServer) readHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	value := s.db.Get(vars["key"])
	w.Write([]byte(value))
}

func (s *RaftServer) writeHandler(w http.ResponseWriter, req *http.Request) {
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
