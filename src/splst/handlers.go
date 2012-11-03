package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"splst/project"
	"strings"
)

var (
	templates = template.Must(template.ParseGlob(path.Join(docRoot, "templates", "*.html")))
)

type handlerError struct {
	Error       error
	Message     interface{}
	Code        int
	ContentType string
}

type splstHandler func(http.ResponseWriter, *http.Request, *sessions.Session) *handlerError

func (f splstHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			log.Print(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}()

	s, _ := genSession(w, r)
	if err := f(w, r, s); err != nil {

		if err.ContentType == "" {
			err.ContentType = "plain/text"
		}

		log.Print(err.Error)

		w.Header().Set("Content-Type", err.ContentType)
		w.WriteHeader(err.Code)

		var message []byte

		switch err.Message.(type) {
		case string:
			message = []byte(err.Message.(string))
		}

		if err.ContentType == "application/json" {
			message, _ = json.Marshal(err.Message)
		}

		w.Write(message)

		return
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) *handlerError {
	vars := mux.Vars(r)
	//key := vars["key"]

	err := templates.ExecuteTemplate(w, "home.html", vars)
	if err != nil {
		return &handlerError{Error: err, Message: "Internal Server Error", Code: http.StatusInternalServerError}
	}

	return nil
}

func addProjectHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) *handlerError {

	errMessage := make(map[string]string)

	projectUrl := strings.TrimSpace(r.FormValue("project-url"))
	projectName := strings.TrimSpace(r.FormValue("project-name"))
	userid := s.Values["userid"].(string)

	if len(projectName) == 0 {
		errMessage["project-name"] = "Project name is requird"
		return &handlerError{Error: errors.New("Project name is requird"), Message: errMessage, Code: http.StatusBadRequest, ContentType: "application/json"}
	}

	if len(projectUrl) == 0 {
		errMessage["project-url"] = "Project URL is requird"
		return &handlerError{Error: errors.New("Project URL is requird"), Message: errMessage, Code: http.StatusBadRequest, ContentType: "application/json"}
	}

	p := project.Project{Name: projectName, URL: projectUrl, OwnerId: userid}
	err := p.Save(projectsRoot)
	if err != nil {
		if err == project.InvalidUrlError {
			errMessage["project-url"] = fmt.Sprintf("%q is not a fully qualified URL", projectUrl)
			return &handlerError{Error: err, Message: errMessage, Code: http.StatusBadRequest, ContentType: "application/json"}
		}

		if err == project.GenerateThumbError {
			errMessage["error"] = "Couldn't generate thumbnail image. Probably there was a problem fetching the URL. Make sure that the submitted URL is correct."
			return &handlerError{Error: err, Message: errMessage, Code: http.StatusInternalServerError, ContentType: "application/json"}
		}

		return &handlerError{Error: err, Message: "Internal Server Error", Code: http.StatusInternalServerError, ContentType: "application/json"}
	}

	return nil
}
