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
	flagHost         = flag.String("host", "", "Hostname to listen on")
	flagPort         = flag.String("port", "9980", "Listening port")
	flagDocRoot      = flag.String("document-root", os.Getenv("PWD"), "Document root containing templates and assets")
	flagProjectsRoot = flag.String("projects-root", path.Join(os.Getenv("PWD"), "static/projects"), "Projects root containing images of projects")

	Usage = func() {

		fmt.Fprintf(os.Stderr, "Usage of monitor:"+
			"\n\t--host            hostname of the monitor (e.g --host localhost)"+
			"\n\t--port            Listening port of the monitor"+
			"\n\t--document-root   Document root containing templates and assets\n")
	}
	docRoot      string
	projectsRoot string
)

func main() {

	runtime.GOMAXPROCS(4)

	flag.Usage = Usage
	flag.Parse()

	docRoot = path.Join(*flagDocRoot, "templates")
	projectsRoot = *flagProjectsRoot
	addr := net.JoinHostPort(*flagHost, *flagPort)

	r := mux.NewRouter()
	r.Handle("/", splstHandler(homeHandler)).Methods("GET")
	r.Handle("/{key}", splstHandler(homeHandler)).Methods("GET")
	r.Handle("/fetch-url-info", splstHandler(fetchURLInfoHandler)).Methods("POST")
	r.Handle("/add-project", splstHandler(addProjectHandler)).Methods("POST")

	http.Handle("/", r)
	http.ListenAndServe(addr, nil)
}
