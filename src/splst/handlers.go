package main

import (
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
	Err     error
	Message interface{}
	Code    int
}

func (e *handlerError) Error() string {
	return e.Err.Error()
}

type splstHandler func(http.ResponseWriter, *http.Request, *sessions.Session) error

func (f splstHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			log.Print(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}()

	s, _ := genSession(w, r)
	var err *handlerError

	if e := f(w, r, s); e != nil {

		// If it's a regular error convert it to *handleError
		if herr, ok := e.(*handlerError); !ok {
			err = &handlerError{Err: e, Message: "Internal Server Error", Code: http.StatusInternalServerError}
		} else {
			err = herr
		}

		contentType := strings.FieldsFunc(r.Header.Get("Accept"), func(sep rune) bool { return ',' == sep })[0]

		if contentType == "" {
			contentType = "plain/text"
		}

		log.Print(err.Err)

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(err.Code)

		var message []byte

		switch err.Message.(type) {
		case string:
			message = []byte(err.Message.(string))
		}

		if contentType == "application/json" {
			message, _ = json.Marshal(err.Message)
		}

		w.Write(message)

		return
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	projects, err := project.RecentList()
	if err != nil {
		return err
	}

	err = templates.ExecuteTemplate(w, "home.html", projects)
	if err != nil {
		return err
	}

	return nil
}

func addProjectHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	errMessage := make(map[string]string)

	projectUrl := strings.TrimSpace(r.FormValue("project-url"))
	projectName := strings.TrimSpace(r.FormValue("project-name"))
	userid := s.Values["userid"].(string)

	if len(projectName) == 0 {
		errMessage["project-name"] = "Project name is requird"
		return &handlerError{Err: errors.New("Project name is requird"), Message: errMessage, Code: http.StatusBadRequest}
	}

	if len(projectUrl) == 0 {
		errMessage["project-url"] = "Project URL is requird"
		return &handlerError{Err: errors.New("Project URL is requird"), Message: errMessage, Code: http.StatusBadRequest}
	}

	p := project.Project{Name: projectName, URL: projectUrl, OwnerId: userid}
	err := p.Save(projectsRoot)
	if err != nil {
		if err == project.InvalidUrlError {
			errMessage["project-url"] = fmt.Sprintf("%q is not a fully qualified URL", projectUrl)
			return &handlerError{Err: err, Message: errMessage, Code: http.StatusBadRequest}
		}

		if err == project.GenerateThumbError {
			errMessage["error"] = "Couldn't generate thumbnail image. Probably there was a problem fetching the URL. Make sure that the submitted URL is correct."
			return &handlerError{Err: err, Message: errMessage, Code: http.StatusInternalServerError}
		}

		return &handlerError{Err: err, Message: "Internal Server Error", Code: http.StatusInternalServerError}
	}

	return nil
}
