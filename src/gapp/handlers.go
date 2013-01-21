package gapp

import (
	"github.com/gorilla/sessions"

	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type HandlerError struct {
	Err     error
	Message interface{}
	Code    int
}

func (e *HandlerError) Error() string {
	return e.Err.Error()
}

type Handler func(http.ResponseWriter, *http.Request, *sessions.Session) error

func (f Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			log.Print(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}()

	s, e := genSession(w, r)
	if e != nil {
		log.Print(e)
	}

	var err *HandlerError

	if e := f(w, r, s); e != nil {

		// If it's a regular error convert it to *handleError
		if herr, ok := e.(*HandlerError); !ok {
			err = &HandlerError{Err: e, Message: "Internal Server Error", Code: http.StatusInternalServerError}
		} else {
			err = herr
		}

		contentType := strings.FieldsFunc(r.Header.Get("Accept"), func(sep rune) bool { return ',' == sep })[0]

		if contentType == "" {
			contentType = "plain/html"
		}

		log.Print(err.Err)

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(err.Code)

		var message []byte

		switch err.Message.(type) {
		case string:
			message = []byte(err.Message.(string))
		}

		switch contentType {
		case "application/json":
			message, _ = json.Marshal(err.Message)
		default:
			data := map[string]interface{}{
				"BUILD":    BuildId,
				"title":    http.StatusText(err.Code),
				"keywords": strconv.Itoa(err.Code) + ", " + http.StatusText(err.Code),
				"error":    err.Message,
			}
			err := Templates.ExecuteTemplate(w, "error.html", data)
			if err == nil {
				return
			}
		}

		w.Write(message)

		return
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	data := map[string]interface{}{
		"BUILD":    BuildId,
		"title":    "Home",
		"keywords": "home",
	}

    err := Templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		return err
	}

	return nil
}
