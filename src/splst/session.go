package main

import (
	"splst/user"
	"github.com/gorilla/sessions"

	"log"
	"net/http"
	"time"
)

var (
	store = sessions.NewCookieStore([]byte("something-very-secret"))
)

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
	var userid uint64
	uid, ok := s.Values["userid"]
	if !ok || uid == 0 {
		var err error
		userid, err = nq.GenOne()
		if err != nil {
			return nil, err
		}

		s.Values["userid"] = userid

		user := user.New(userid)
		err = user.Save()
		if err != nil {
			return nil, err
		}

	} else {
		userid = uid.(uint64)
	}

	user := user.New(userid)
	err = user.Fetch()
	if err != nil {
		return nil, err
	}

	user.LastAccess = time.Now()
	err = user.Update()
	if err != nil {
		return nil, err
	}

	log.Printf("%+v", user)

	// Saving session everytime it gets access helps to push expiry date further
	s.Save(r, w)

	return s, nil
}
