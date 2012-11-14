package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
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

	userid := s.Values["userid"].(string)

	projects, err := project.RecentList()
	if err != nil {
		return err
	}

	err = templates.ExecuteTemplate(w, "home.html", map[string]interface{}{"projects": projects, "userid": userid, "recent": true})
	if err != nil {
		return err
	}

	return nil
}

func mineHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	userid := s.Values["userid"].(string)

	projects, err := project.MyList(userid)
	if err != nil {
		return err
	}

	err = templates.ExecuteTemplate(w, "home.html", map[string]interface{}{"projects": projects, "userid": userid, "mine": true})
	if err != nil {
		return err
	}

	return nil
}

func fetchURLInfoHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	url := r.FormValue("url")

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := make([]byte, 4096)
	io.ReadFull(resp.Body, buf)

	reg := regexp.MustCompile("<title>([^>]+)</title>")
	title := reg.FindStringSubmatch(string(buf))

	reg = regexp.MustCompile("name=\"description\"\\s+content=\"([^\"]+)\"")
	desc := reg.FindStringSubmatch(string(buf))

	info := make(map[string]string)

	if len(title) == 2 {
		info["name"] = strings.TrimSpace(title[1])
	}

	if len(desc) == 2 {
		info["description"] = strings.TrimSpace(desc[1])
	}

	result, _ := json.Marshal(info)
	w.Write(result)

	return nil
}

func addProjectHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	errMessage := make(map[string]string)

	projectUrl := strings.TrimSpace(r.FormValue("url"))
	projectName := strings.TrimSpace(r.FormValue("name"))
	projectDescription := strings.TrimSpace(r.FormValue("description"))
	projectRepository := strings.TrimSpace(r.FormValue("code-repo"))
	userid := s.Values["userid"].(string)

	if len(projectName) == 0 {
		errMessage["name"] = "Project name is requird"
		return &handlerError{Err: errors.New("Project name is requird"), Message: errMessage, Code: http.StatusBadRequest}
	}

	if len(projectUrl) == 0 {
		errMessage["url"] = "Project URL is requird"
		return &handlerError{Err: errors.New("Project URL is requird"), Message: errMessage, Code: http.StatusBadRequest}
	}

	_, err := url.ParseRequestURI(projectUrl)
	if err != nil {
		errMessage["url"] = fmt.Sprintf("%q is not a fully qualified URL", projectUrl)
		return &handlerError{Err: err, Message: errMessage, Code: http.StatusBadRequest}
	}

	go func() {
		p := &project.Project{Name: projectName, URL: projectUrl, OwnerId: userid, Description: projectDescription, RepositoryURL: projectRepository}
		err := p.Save(projectsRoot)
		if err != nil {
			log.Printf("Error in saving project %q by user %q - %s", p.Id, p.OwnerId, err)
		}
	}()

	return nil
}

func deleteProjectHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {
	vars := mux.Vars(r)
	pid := vars["pid"]

	p, err := project.Fetch(pid)
	if err != nil {
		return err
	}

	return p.Delete(projectsRoot)
}
