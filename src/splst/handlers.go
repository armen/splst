package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"html/template"
	"log"
	"net/http"
	"path"
)

var (
	templates = template.Must(template.ParseGlob(path.Join(docRoot, "templates", "*.html")))
)

type splstHandler func(http.ResponseWriter, *http.Request, *sessions.Session) error

func (f splstHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			log.Print(err)
			http.Error(w, "{success:false}", http.StatusServiceUnavailable)
		}
	}()

	s, _ := genSession(w, r)
	if err := f(w, r, s); err != nil {
		log.Printf("%s", err)
		http.Error(w, "{success:false}", http.StatusServiceUnavailable)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {
	vars := mux.Vars(r)
	//key := vars["key"]

	err := templates.ExecuteTemplate(w, "home.html", vars)

	return err
}
