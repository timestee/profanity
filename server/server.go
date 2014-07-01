package server

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/simonz05/profanity/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/simonz05/profanity/db"
	"github.com/simonz05/profanity/Godeps/_workspace/src/github.com/simonz05/util/log"
)

var (
	Version = "0.1.0"
	router  *mux.Router
	filters *profanityFilters
	dbConn  db.Conn
)

func sigTrapCloser(l net.Listener) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for _ = range c {
			l.Close()
			log.Printf("Closed listener %s", l.Addr())
		}
	}()
}

func setupServer(dsn string) (err error) {
	dbConn, err = db.Open(dsn)

	if err != nil {
		return
	}

	filters = newProfanityFilters()

	// HTTP endpoints
	router = mux.NewRouter()
	router.HandleFunc("/v1/profanity/sanitize/", sanitizeHandle).Methods("GET").Name("sanitize")
	router.HandleFunc("/v1/profanity/blacklist/", updateBlacklistHandle).Methods("POST", "PUT").Name("blacklist")
	router.HandleFunc("/v1/profanity/blacklist/remove/", removeBlacklistHandle).Methods("POST", "PUT").Name("blacklist")
	router.HandleFunc("/v1/profanity/blacklist/", getBlacklistHandle).Methods("GET").Name("blacklist")
	router.StrictSlash(false)
	http.Handle("/", router)
	return
}

func ListenAndServe(laddr, dsn string) error {
	setupServer(dsn)

	l, err := net.Listen("tcp", laddr)

	if err != nil {
		return err
	}

	log.Printf("Listen on %s", l.Addr())

	sigTrapCloser(l)
	err = http.Serve(l, nil)
	log.Print("Shutting down ..")
	return err
}
