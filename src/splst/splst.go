package main

import (
	"github.com/gorilla/mux"

	"flag"
	"fmt"
	"gapp"
	"net/http"
	"os"
	"splst/project"
)

var (
	flagConf = flag.String("conf", "conf/app.ini", "Configuration file")
	Usage    = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n\t--conf Configuration file (e.g --conf conf/app.ini)\n")
	}
)

func main() {

	flag.Usage = Usage
	flag.Parse()

	gapp.Init(*flagConf)

	saveConcurrencySize, err := gapp.Config.GetInt("default", "save-concurrency-size")
	if err != nil {
		saveConcurrencySize = 2
	}

	project.Init(gapp.RedisPool, gapp.AppRoot, saveConcurrencySize)

	r := mux.NewRouter()
	r.Handle("/", gapp.Handler(homeHandler)).Methods("GET")
	r.Handle("/recent", gapp.Handler(homeHandler)).Methods("GET")
	r.Handle("/mine", gapp.Handler(mineHandler)).Methods("GET")
	r.Handle("/url-info", gapp.Handler(fetchURLInfoHandler)).Methods("GET")
	r.Handle("/project", gapp.Handler(addProjectHandler)).Methods("POST")
	r.Handle("/project/{pid}", gapp.Handler(deleteProjectHandler)).Methods("DELETE")
	r.Handle("/project/{pid}", gapp.Handler(projectHandler)).Methods("GET")
	r.Handle("/signin", gapp.SigninHandler).Methods("GET")
	r.Handle("/google-signin", gapp.GoogleSigninHandler).Methods("POST")
	r.Handle("/google-callback", gapp.GoogleCallbackHandler).Methods("GET")
	r.Handle("/{page}", gapp.PageHandler).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(gapp.Address, nil)
}
