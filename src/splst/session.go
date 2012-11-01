package main

import (
	"github.com/gorilla/sessions"

	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
)

var (
	store = sessions.NewCookieStore([]byte("something-very-secret"))
)

func id() string {
	buf := make([]byte, 16)
	io.ReadFull(rand.Reader, buf)
	return fmt.Sprintf("%x", buf)
}

func genSession(w http.ResponseWriter, r *http.Request) (*sessions.Session, error) {
	// Create a splst session and store it in cookie so that we can recognize the user when he/she gets back
	s, err := store.Get(r, "splst")
	if err != nil {
		return nil, err
	}

	// Changed maximum age of the session to one month
	s.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 30,
	}

	// Generate new userid if there isn't any
	userid, ok := s.Values["userid"]
	if !ok {
		userid = id()
		s.Values["userid"] = userid
	}

	log.Printf("%+v", userid)

	// Saving session everytime it gets access helps to push expiry date further
	s.Save(r, w)

	return s, nil
}
