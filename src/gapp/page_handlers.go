package gapp

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"fmt"
	"net/http"
	"path"
	"strings"
)

func pageHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {
	vars := mux.Vars(r)
	page := vars["page"]

	switch page {
	case "about", "privacy", "feedback":

		data := map[string]interface{}{
			"BUILD":    BuildId,
			"page":     map[string]bool{page: true}, // Select the page in the top navbar
			"title":    strings.Title(page),
			"keywords": page,
		}

		err := Templates.ExecuteTemplate(w, path.Join(page+".html"), data)
		if err != nil {
			return err
		}
	default:
		err := fmt.Errorf("Page %q could not be found!", page)
		return &HandlerError{Err: err, Message: err.Error(), Code: http.StatusNotFound}
	}

	return nil
}
