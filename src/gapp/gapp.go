package gapp

import (
	"code.google.com/p/goconf/conf"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/sessions"

	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"runtime"
	"time"
)

var (
	Config    conf.ConfigFile
	DocRoot   string
	AppRoot   string
	Hostname  string
	Host      string
	Port      string
	Address   string
	RedisPool *redis.Pool
	Templates *template.Template
	BuildId   string
)

var (
	SigninHandler         = Handler(signinHandler)
	GoogleSigninHandler   = Handler(googleSigninHandler)
	GoogleCallbackHandler = Handler(googleCallbackHandler)
	PageHandler           = Handler(pageHandler)
)

var (
	sessionSecrets [][]byte
	sessionStore   *sessions.CookieStore
)

func Init(configFile string) {

	Config, err := conf.ReadConfigFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	goMaxProcs, err := Config.GetInt("default", "go-max-procs")
	if err != nil {
		goMaxProcs = 3
	}

	AppRoot, err = Config.GetString("default", "app-root")
	if err != nil {
		AppRoot = os.Getenv("PWD")
	}

	Hostname, err = Config.GetString("default", "hostname")
	if err != nil {
		Hostname = "localhost:8080"
	}

	Host, err := Config.GetString("default", "host")
	if err != nil {
		Host = "localhost"
	}

	Port, err := Config.GetString("default", "port")
	if err != nil {
		Port = "9980"
	}

	secrets, err := Config.GetString("session", "secrets")
	if err != nil {
		log.Fatal(err)
	}

	redisMaxIdle, err := Config.GetInt("redis", "max-idle")
	if err != nil {
		redisMaxIdle = 20
	}

	redisIdleTimeout, err := Config.GetInt("redis", "idle-timeout")
	if err != nil {
		redisIdleTimeout = 240
	}

	build, err := ioutil.ReadFile(path.Join(AppRoot, "conf", "BUILD"))
	if err != nil {
		log.Fatalf("%s - Please make sure BUILD file is created with \"make styles\"", err)
	}
	BuildId = string(bytes.TrimSpace(build))

	RedisPool = &redis.Pool{
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
	DocRoot = path.Join(AppRoot, "templates")
	Address = net.JoinHostPort(Host, Port)
	sessionStore = sessions.NewCookieStore(bytes.Fields([]byte(secrets))...)
	Templates = template.Must(template.ParseGlob(path.Join(DocRoot, "*.html")))
}
