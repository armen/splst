package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path"
	"splst/project"
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

func addProjectHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	projectUrl := r.FormValue("project-url")
	projectName := r.FormValue("project-name")
	userid := s.Values["userid"].(string)

	p := project.Project{Name: projectName, URL: projectUrl, OwnerId: userid}
	err := p.Save()
	if err != nil {
		return err
	}

	v, _ := json.Marshal(nil)
	w.Write(v)

	return nil
}
