package main

import (
	"code.google.com/p/goconf/conf"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
	"splst/project"
	"time"
)

var (
	config         conf.ConfigFile
	docRoot        string
	appRoot        string
	sessionSecrets [][]byte
	templates      *template.Template
	store          *sessions.CookieStore

	flagConf = flag.String("conf", "conf/splst.ini", "Configuration file")
	Usage    = func() {
		fmt.Fprintf(os.Stderr, "Usage of splst:\n\t--conf Configuration file (e.g --conf conf/splst.ini)\n")
	}
)

func main() {

	flag.Usage = Usage
	flag.Parse()

	config, err := conf.ReadConfigFile(*flagConf)
	if err != nil {
		log.Fatal(err)
	}

	goMaxProcs, err := config.GetInt("default", "go-max-procs")
	if err != nil {
		goMaxProcs = 3
	}

	appRoot, err := config.GetString("default", "app-root")
	if err != nil {
		appRoot = os.Getenv("PWD")
	}

	host, err := config.GetString("default", "host")
	if err != nil {
		host = "localhost"
	}

	port, err := config.GetString("default", "port")
	if err != nil {
		port = "9980"
	}

	saveConcurrencySize, err := config.GetInt("default", "save-concurrency-size")
	if err != nil {
		saveConcurrencySize = 2
	}

	secrets, err := config.GetString("session", "secrets")
	if err != nil {
		log.Fatal(err)
	}

	redisMaxIdle, err := config.GetInt("redis", "max-idle")
	if err != nil {
		redisMaxIdle = 20
	}

	redisIdleTimeout, err := config.GetInt("redis", "idle-timeout")
	if err != nil {
		redisIdleTimeout = 240
	}

	redisPool := &redis.Pool{
		MaxIdle:     redisMaxIdle,
		IdleTimeout: time.Duration(redisIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}

	runtime.GOMAXPROCS(goMaxProcs)
	docRoot = path.Join(appRoot, "templates")
	addr := net.JoinHostPort(host, port)
	store = sessions.NewCookieStore(bytes.Fields([]byte(secrets))...)
	templates = template.Must(template.ParseGlob(path.Join(docRoot, "*.html")))

	project.Init(redisPool, appRoot, saveConcurrencySize)

	r := mux.NewRouter()
	r.Handle("/", splstHandler(homeHandler)).Methods("GET")
	r.Handle("/recent", splstHandler(homeHandler)).Methods("GET")
	r.Handle("/mine", splstHandler(mineHandler)).Methods("GET")
	r.Handle("/url-info", splstHandler(fetchURLInfoHandler)).Methods("GET")
	r.Handle("/project", splstHandler(addProjectHandler)).Methods("POST")
	r.Handle("/project/{pid}", splstHandler(deleteProjectHandler)).Methods("DELETE")
	r.Handle("/project/{pid}", splstHandler(projectHandler)).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(addr, nil)
}
