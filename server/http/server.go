package http

import (
	//"fmt"
	//"io"
	//"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	log "github.com/uthng/golog"

	httphandler "github.com/uthng/container-injector/handlers/http"
)

// Server describes a list of functions for this interface
//type Server interface {
//Serve() error
//}

// Server describes server's elements
type Server struct {
	logger *log.Logger

	mutate http.Handler

	addr     string
	certFile string
	keyFile  string
}

// NewServer returns a new interface
func NewServer(addr, certFile, keyFile string, logger *log.Logger) *Server {
	s := &Server{
		logger:   logger,
		addr:     addr,
		certFile: certFile,
		keyFile:  keyFile,
	}

	s.mutate = httphandler.NewMutate(logger)

	return s
}

// Serve launches http server
func (s *Server) Serve() error {
	r := mux.NewRouter()

	r.Handle("/mutate", s.mutate).Methods("POST")
	r.HandleFunc("/health/ready", s.handleReady).Methods("GET")
	http.Handle("/", accessControl(r))

	return http.ListenAndServeTLS(s.addr, s.certFile, s.keyFile, nil)
}

///////////// INTERNAL FUNCTIONS /////////////////

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	// If we reached this point it means we served a TLS certificate.
	w.WriteHeader(204)
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
