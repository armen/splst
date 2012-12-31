package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"splst/project"
	"strings"
)

func homeHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	userid := s.Values["userid"].(string)

	projects, err := project.RecentList()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"BUILD":           string(BUILD),
		"projects":        projects,
		"userid":          userid,
		"projectPage":     map[string]bool{"recent": true},
		"title":           "Recent Projects",
		"keywords":        "recent projects, latest projects, new projects",
		"newcomer":        !project.HasList(userid),
		"myProjectsCount": project.ListCount(userid),
		"jobsCount":       project.JobsCount(userid),
	}

	err = templates.ExecuteTemplate(w, "home.html", data)
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

	data := map[string]interface{}{
		"BUILD":           string(BUILD),
		"projects":        projects,
		"userid":          userid,
		"projectPage":     map[string]bool{"mine": true},
		"title":           "My Projects",
		"keywords":        "my projects, add projects",
		"newcomer":        !project.HasList(userid),
		"myProjectsCount": project.ListCount(userid),
		"jobsCount":       project.JobsCount(userid),
	}

	err = templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		return err
	}

	return nil
}

func fetchURLInfoHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	projectURL := r.URL.Query().Get("url")

	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(projectURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := make([]byte, 4096)
	io.ReadFull(resp.Body, buf)

	reg := regexp.MustCompile("<\\s*title\\s*>([^<]+)")
	title := reg.FindStringSubmatch(string(buf))

	// First find the description meta tag then extract its content
	reg = regexp.MustCompile(".*name\\s*=.*[ '\"]description[^>]+")
	descmeta := reg.FindStringSubmatch(string(buf))

	var desc []string
	if len(descmeta) > 0 {
		reg = regexp.MustCompile("content\\s*=\\s*['\"]([^'\"]+)['\"]")
		desc = reg.FindStringSubmatch(string(descmeta[0]))
	}

	// First find the link with icon rel then extract href, doing it in two steps helps to parse
	// links with rel or href at the begining
	reg = regexp.MustCompile("link.*rel\\s*=.*[ '\"]icon[^>]+")
	faviconlink := reg.FindStringSubmatch(string(buf))

	var fav []string
	if len(faviconlink) > 0 {
		reg = regexp.MustCompile("href\\s*=\\s*['\"]([^'\"]+)['\"]")
		fav = reg.FindStringSubmatch(string(faviconlink[0]))
	}

	info := make(map[string]string)

	if len(title) == 2 {
		info["name"] = strings.TrimSpace(title[1])
	}

	if len(desc) == 2 {
		info["description"] = strings.TrimSpace(desc[1])
	}

	var faviconURL *url.URL
	if len(fav) == 2 {
		faviconURL, _ = url.Parse(fav[1])
	} else {
		faviconURL, _ = url.Parse("/favicon.ico")
	}

	purl, _ := url.Parse(projectURL)
	favicon := purl.ResolveReference(faviconURL).String()

	if res, err := client.Head(favicon); err == nil && res.StatusCode == 200 {
		info["favicon"] = favicon
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
	projectFavicon := strings.TrimSpace(r.FormValue("favicon"))
	userid := s.Values["userid"].(string)

	log.Println("armen", projectFavicon)

	if len(projectName) == 0 {
		errMessage["name"] = "Project name is required"
		return &handlerError{Err: errors.New("Project name is required"), Message: errMessage, Code: http.StatusBadRequest}
	}

	if len(projectUrl) == 0 {
		errMessage["url"] = "Project URL is required"
		return &handlerError{Err: errors.New("Project URL is required"), Message: errMessage, Code: http.StatusBadRequest}
	}

	_, err := url.ParseRequestURI(projectUrl)
	if err != nil {
		errMessage["url"] = fmt.Sprintf("%q is not a fully qualified URL", projectUrl)
		return &handlerError{Err: err, Message: errMessage, Code: http.StatusBadRequest}
	}

	go func() {
		p := &project.Project{Name: projectName, URL: projectUrl, OwnerId: userid, Description: projectDescription, RepositoryURL: projectRepository, Favicon: projectFavicon}
		err := p.Save()
		if err != nil {
			log.Printf("Error in saving project %q by user %q - %s", p.Id, p.OwnerId, err)
		}
	}()

	return nil
}

func deleteProjectHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {
	vars := mux.Vars(r)
	pid := vars["pid"]
	userid := s.Values["userid"].(string)

	p, err := project.Fetch(pid)
	if err != nil {
		return &handlerError{Err: err, Message: "Not Found", Code: http.StatusNotFound}
	}

	if p.Mine(userid) {
		return p.Delete()
	}

	return &handlerError{Err: errors.New("Permission Denied"), Message: "Permission Denied", Code: http.StatusForbidden}
}

func projectHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {
	vars := mux.Vars(r)
	pid := vars["pid"]
	userid := s.Values["userid"].(string)

	p, err := project.Fetch(pid)
	if err != nil {
		return &handlerError{Err: err, Message: "Not Found", Code: http.StatusNotFound}
	}

	data := map[string]interface{}{
		"BUILD":           string(BUILD),
		"project":         p,
		"userid":          userid,
		"projectPage":     map[string]bool{"detail": true},
		"title":           p.Name,
		"keywords":        "project detail",
		"myProjectsCount": project.ListCount(userid),
		"jobsCount":       project.JobsCount(userid),
	}

	err = templates.ExecuteTemplate(w, "project.html", data)
	if err != nil {
		return err
	}

	return nil
}
