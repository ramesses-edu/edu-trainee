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
)

var (
	oauthFacebook *oauth2.Config = &oauth2.Config{
		ClientID:     "346852496742371",
		ClientSecret: "aa8338431fe3a7cfa79d9abce2812b95",
		RedirectURL:  "http://localhost:80/auth/callback/facebook",
		Scopes:       []string{"public_profile", "email"},
		Endpoint:     facebookEndpoint,
	}
	facebookEndpoint oauth2.Endpoint = oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("https://www.facebook.com/%s/dialog/oauth", facebookAPIVersion),
		TokenURL: fmt.Sprintf("https://graph.facebook.com/%s/oauth/access_token", facebookAPIVersion),
	}
	facebookAPIVersion = "v10.0"
	oauthStateFaceBook = ""
)

func authFacebook(w http.ResponseWriter, r *http.Request) {
	Url, err := url.Parse(oauthFacebook.Endpoint.AuthURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	oauthStateFaceBook = generateOauthStateProvider()
	cookie := http.Cookie{Name: "oauthstate", Value: oauthStateFaceBook, Expires: time.Now().Add(5 * time.Minute)}
	http.SetCookie(w, &cookie)
	////////////////////////////////////////////////////
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
	if state != (oauthstate.Value) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	///////////////////////////////////////////////////////////
	code := r.FormValue("code")
	token, err := oauthFacebook.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	vals := url.Values{}
	vals.Add("fields", "id,name,email")
	vals.Add("access_token", url.QueryEscape(token.AccessToken))
	resp, err := http.Get(fmt.Sprintf("https://graph.facebook.com/%s/me?%s", facebookAPIVersion, vals.Encode()))
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
	//decode answer JSON to map
	var respMap map[string]interface{} = make(map[string]interface{})
	err = json.Unmarshal(response, &respMap)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	//check request error
	if _, ok := respMap["error"]; ok {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	//generate new accessToken for user
	accessToken := generateAccessToken()
	hashAccToken := calculateSignature(accessToken, "provider")
	//check user registration
	var u user
	result := u.getUser(a.db, map[string]interface{}{
		"login":    respMap["id"],
		"provider": "facebook",
	})
	//if user not found, register new user
	if result.Error != nil || result.RowsAffected == 0 {
		u = user{
			Login:       respMap["id"].(string),
			Provider:    "facebook",
			Name:        respMap["name"].(string),
			AccessToken: hashAccToken,
		}
		result = u.createUser(a.db)
	} else {
		u.AccessToken = hashAccToken
		u.updateAccessToken(a.db)
	}
	//write cookies
	if result.Error == nil {
		var expiration = time.Now().Add(30 * 24 * time.Hour)
		cookieUID := http.Cookie{Name: "UAAT", Value: accessToken, Expires: expiration, Path: "/"}
		http.SetCookie(w, &cookieUID)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
