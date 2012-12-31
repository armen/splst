package main

import (
	"code.google.com/p/goauth2/oauth"
	"github.com/gorilla/sessions"

	"net/http"
)

var oauthCfg = &oauth.Config{
	ClientId:     "507340711959-ltgjmi1eg2sdeqdcn3ci564svu1frcr1.apps.googleusercontent.com",
	ClientSecret: "jELmHlwDLRKXVwovPw4f2LAR",
	Scope:        "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/plus.me",
	AuthURL:      "https://accounts.google.com/o/oauth2/auth",
	TokenURL:     "https://accounts.google.com/o/oauth2/token",
	RedirectURL:  "http://localhost:8080/oauth2callback",
}

func signinHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	data := map[string]interface{}{
		"BUILD":    string(BUILD),
		"page":     map[string]bool{"signin": true}, // Select signin in the top navbar
		"title":    "Signin & Signup",
		"keywords": "signin, signup",
	}

	err := templates.ExecuteTemplate(w, "signin.html", data)
	if err != nil {
		return err
	}

	return nil
}

func googleSigninHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {
	// Get the Google URL which shows the Authentication page to the user
	url := oauthCfg.AuthCodeURL("")

	// Redirect user to that page
	http.Redirect(w, r, url, http.StatusFound)

	return nil
}
