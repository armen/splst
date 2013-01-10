package main

import (
	"code.google.com/p/goauth2/oauth"
	"github.com/gorilla/sessions"

	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type User struct {
	ID            string
	OauthID       string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email",bool`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Link          string `json:"link"`
	Gender        string `json:"gender"`
	Locale        string `json:"locale"`
	HostDomain    string `json:"hd"`
}

var oauthCfg = &oauth.Config{
	ClientId:     "507340711959-ltgjmi1eg2sdeqdcn3ci564svu1frcr1.apps.googleusercontent.com",
	ClientSecret: "jELmHlwDLRKXVwovPw4f2LAR",
	Scope:        "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile",
	AuthURL:      "https://accounts.google.com/o/oauth2/auth",
	TokenURL:     "https://accounts.google.com/o/oauth2/token",
}

const (
	redirectURL    = "http://%s/google-callback"
	profileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo?alt=json"
)

func signinHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	data := map[string]interface{}{
		"BUILD":    string(BUILD),
		"page":     map[string]bool{"signin": true}, // Select signin in the top navbar
		"title":    "Signin & Signup",
		"keywords": "signin, signup, loging, register",
	}

	err := templates.ExecuteTemplate(w, "signin.html", data)
	if err != nil {
		return err
	}

	return nil
}

func googleSigninHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	// Replace %s with hostname
	oauthCfg.RedirectURL = fmt.Sprintf(redirectURL, strings.TrimSpace(splstHostname))

	// Get the Google URL which shows the Authentication page to the user
	url := oauthCfg.AuthCodeURL("")

	// Redirect user to that page
	http.Redirect(w, r, url, http.StatusFound)

	return nil
}

func googleCallbackHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) error {

	userid := s.Values["userid"].(string)

	// Get the code from the response
	code := r.FormValue("code")

	t := &oauth.Transport{oauth.Config: oauthCfg}

	// Exchange the received code for a token
	t.Exchange(code)

	// Now get user data based on the Transport which has the token
	resp, err := t.Client().Get(profileInfoURL)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	c := splstRedisPool.Get()
	defer c.Close()

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return err
	}

	user.ID = userid
	fmt.Fprintf(w, "%#v", user)

	return nil
}
