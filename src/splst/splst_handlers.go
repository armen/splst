package main

import (
	"github.com/gorilla/sessions"

	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	s, e := genSession(w, r)
	if e != nil {
		log.Print(e)
	}

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
				"BUILD":    string(BUILD),
				"title":    http.StatusText(err.Code),
				"keywords": strconv.Itoa(err.Code) + ", " + http.StatusText(err.Code),
				"error":    err.Message,
			}
			err := templates.ExecuteTemplate(w, "error.html", data)
			if err == nil {
				return
			}
		}

		w.Write(message)

		return
	}
}
