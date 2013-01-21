# GAPP

## First application with GAPP

Create a directory in src/ and name your application (e.g. app)

    mkdir src/app

Then create it's main file (e.g. src/app/app.go)

```go
package main

import (
    "github.com/gorilla/mux"

    "flag"
    "fmt"
    "gapp"
    "log"
    "net/http"
    "os"
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

    r := mux.NewRouter()
    r.Handle("/", gapp.Handler(gapp.HomeHandler)).Methods("GET")
    r.Handle("/signin", gapp.SigninHandler).Methods("GET")
    r.Handle("/google-signin", gapp.GoogleSigninHandler).Methods("POST")
    r.Handle("/google-callback", gapp.GoogleCallbackHandler).Methods("GET")
    r.Handle("/{page}", gapp.PageHandler).Methods("GET")

    http.Handle("/", r)
    err := http.ListenAndServe(gapp.Address, nil)

    if err != nil {
        log.Fatal(err)
    }
}
```

Now build the application (e.g app)

    make
    go get app

And run it

    ./bin/app

