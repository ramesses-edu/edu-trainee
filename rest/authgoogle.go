package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	oauthGoogle *oauth2.Config = &oauth2.Config{
		ClientID:     "89020002530-f3kmq3qc5me63q58giisrr3lsjguhbre.apps.googleusercontent.com",
		ClientSecret: "kI0AfuN6k2d_AzMZySHmoaXS",
		RedirectURL:  "http://localhost/auth/callback/google",
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile", "openid"},
		Endpoint: google.Endpoint,
	}
	oauthStateGoogle = ""
)

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(24 * time.Hour)
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)
	return state
}

func authGoogle(w http.ResponseWriter, r *http.Request) {
	URL, err := url.Parse(oauthGoogle.Endpoint.AuthURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	oauthStateGoogle = generateStateOauthCookie(w)
	parameters := url.Values{}
	parameters.Add("client_id", oauthGoogle.ClientID)
	parameters.Add("scope", strings.Join(oauthGoogle.Scopes, " "))
	parameters.Add("redirect_uri", oauthGoogle.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", oauthStateGoogle)
	URL.RawQuery = parameters.Encode()
	url := URL.String()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func callbackGoogle(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	oauthstate, _ := r.Cookie("oauthstate")
	if state != oauthstate.Value {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	code := r.FormValue("code")
	if code == "" {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	token, err := oauthGoogle.Exchange(context.Background(), code)
	if err != nil {
		return
	}
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	// w.Write([]byte("Hello, I'm protected\n"))
	// w.Write([]byte(string(response)))
	/*
					Hello, I'm protected
				{
				  "id": "111918739891765423859",
				  "email": "romgrishin@gmail.com",
				  "verified_email": true,
				  "name": "Roman Grishin",
				  "given_name": "Roman",
				  "family_name": "Grishin",
				  "picture": "https://lh6.googleusercontent.com/-snn4x5qPBD8/AAAAAAAAAAI/AAAAAAAAAAA/AMZuuck2NJAT99ofWxbbVePWYIjzzQczcw/s96-c/photo.jpg",
				  "locale": "ru"
				}
				{
		  "error": {
		    "code": 401,
		    "message": "Request is missing required authentication credential. Expected OAuth 2 access token, login cookie or other valid authentication credential. See https://developers.google.com/identity/sign-in/web/devconsole-project.",
		    "status": "UNAUTHENTICATED"
			if map[error] != nil
		}
		}
	*/
	//write cookies
	var responseMap map[string]interface{} = make(map[string]interface{})
	err = json.Unmarshal(response, &responseMap)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	var expiration = time.Now().Add(30 * 24 * time.Hour)
	cookieProvider := http.Cookie{Name: "provider", Value: "google", Expires: expiration, Path: "/"}
	cookieToken := http.Cookie{Name: "TASID", Value: token.AccessToken, Expires: expiration, Path: "/"}
	//cookieUID := http.Cookie{Name: "SAUID", Value: responseMap["id"].(string), Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookieProvider)
	http.SetCookie(w, &cookieToken)
	//http.SetCookie(w, &cookieUID)
}
