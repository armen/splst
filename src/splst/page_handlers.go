package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"fmt"
	"net/http"
	"strings"
)

func pageHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {
	vars := mux.Vars(r)
	page := vars["page"]

	switch page {
	case "about", "privacy", "feedback":

		data := map[string]interface{}{
			"BUILD":    string(BUILD),
			"page":     map[string]bool{page: true},
			"title":    strings.Title(page),
			"keywords": page,
		}

		err := templates.ExecuteTemplate(w, page+".html", data)
		if err != nil {
			return err
		}
	default:
		err := fmt.Errorf("Page %q could not be found!", page)
		return &handlerError{Err: err, Message: err.Error(), Code: http.StatusNotFound}
	}

	return nil
}
