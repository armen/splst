package main

import (
	"github.com/gorilla/mux"

	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
)

var (
	flagHost    = flag.String("host", "", "Hostname to listen on")
	flagPort    = flag.String("port", "9980", "Listening port")
	flagDocRoot = flag.String("document-root", os.Getenv("PWD"), "Document root containing templates and assets")
	flagAppRoot = flag.String("application-root", os.Getenv("PWD"), "Application root")

	Usage = func() {

		fmt.Fprintf(os.Stderr, "Usage of monitor:"+
			"\n\t--host            hostname of the monitor (e.g --host localhost)"+
			"\n\t--port            Listening port of the monitor"+
			"\n\t--document-root   Document root containing templates and assets\n")
	}
	docRoot string
	appRoot string
)

func main() {

	runtime.GOMAXPROCS(3)

	flag.Usage = Usage
	flag.Parse()

	docRoot = path.Join(*flagDocRoot, "templates")
	appRoot = *flagAppRoot
	addr := net.JoinHostPort(*flagHost, *flagPort)

	r := mux.NewRouter()
	r.Handle("/", splstHandler(homeHandler)).Methods("GET")
	r.Handle("/recent", splstHandler(homeHandler)).Methods("GET")
	r.Handle("/mine", splstHandler(mineHandler)).Methods("GET")
	r.Handle("/url-info", splstHandler(fetchURLInfoHandler)).Methods("GET")
	r.Handle("/project", splstHandler(addProjectHandler)).Methods("POST")
	r.Handle("/project/{pid}", splstHandler(deleteProjectHandler)).Methods("DELETE")

	http.Handle("/", r)
	http.ListenAndServe(addr, nil)
}
