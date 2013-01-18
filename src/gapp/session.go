package gapp

import (
	"github.com/gorilla/sessions"

	"gapp/utils"
	"net/http"
)

func genSession(w http.ResponseWriter, r *http.Request) (*sessions.Session, error) {
	// Create a session and store it in cookie so that we can recognize the user when he/she gets back
	// TODO: read session/cookie name from config
	s, err := sessionStore.Get(r, "gapp")
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
		userid = utils.GenId(16)
		s.Values["userid"] = userid
	}

	// Saving session everytime it gets access helps to push expiry date further
	err = s.Save(r, w)

	return s, err
}
