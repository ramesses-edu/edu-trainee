package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

var (
	oauthFacebook *oauth2.Config = &oauth2.Config{
		ClientID:     "346852496742371",
		ClientSecret: "aa8338431fe3a7cfa79d9abce2812b95",
		RedirectURL:  "http://localhost:80/auth/callback/facebook",
		Scopes:       []string{"public_profile", "email", "user_friends"},
		Endpoint:     facebook.Endpoint,
	}
	oauthStateFaceBook = ""
)

func authFacebook(w http.ResponseWriter, r *http.Request) {
	Url, err := url.Parse(oauthFacebook.Endpoint.AuthURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	oauthStateFaceBook = generateStateOauthCookie(w)
	parameters := url.Values{}
	parameters.Add("client_id", oauthFacebook.ClientID)
	parameters.Add("scope", strings.Join(oauthFacebook.Scopes, " "))
	parameters.Add("redirect_uri", oauthFacebook.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", oauthStateFaceBook)
	Url.RawQuery = parameters.Encode()
	url := Url.String()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func callbackFacebook(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	oauthstate, err := r.Cookie("oauthstate")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if state != oauthstate.Value {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateFaceBook, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")

	token, err := oauthFacebook.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://graph.facebook.com/me?access_token=" +
		url.QueryEscape(token.AccessToken))
	if err != nil {
		fmt.Printf("Get: %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ReadAll: %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	/*
					Hello, I'm protected
				{
					"name":"\u0420\u043e\u043c\u0430\u043d \u0413\u0440\u0438\u0448\u0438\u043d",
					"id":"114187967424535"
				}
				{
		   "error": {
		      "message": "The access token could not be decrypted",
		      "type": "OAuthException",
		      "code": 190,
		      "fbtrace_id": "AN1DggkkPDTYaISxwWzqkiR"
		   }
		}
	*/

	//	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	// w.Write([]byte("Hello, I'm protected\n"))
	// w.Write([]byte(string(response)))
	//write cookies
	var responseMap map[string]interface{} = make(map[string]interface{})
	err = json.Unmarshal(response, &responseMap)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	var expiration = time.Now().Add(30 * 24 * time.Hour)
	cookieProvider := http.Cookie{Name: "provider", Value: "facebook", Expires: expiration, Path: "/"}
	cookieToken := http.Cookie{Name: "TASID", Value: token.AccessToken, Expires: expiration, Path: "/"}
	//cookieUID := http.Cookie{Name: "SAUID", Value: responseMap["id"].(string), Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookieProvider)
	http.SetCookie(w, &cookieToken)
	//http.SetCookie(w, &cookieUID)
}
