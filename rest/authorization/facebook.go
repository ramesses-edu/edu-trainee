package authorization

import (
	"context"
	"edu-trainee/rest/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func AuthFacebook(w http.ResponseWriter, r *http.Request) {
	Url, err := url.Parse(A.OauthFacebook.Endpoint.AuthURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	oauthStateFaceBook := generateOauthStateProvider()
	cookie := http.Cookie{Name: "oauthstate", Value: oauthStateFaceBook, Expires: time.Now().Add(5 * time.Minute)}
	http.SetCookie(w, &cookie)
	////////////////////////////////////////////////////
	parameters := url.Values{}
	parameters.Add("client_id", A.OauthFacebook.ClientID)
	parameters.Add("scope", strings.Join(A.OauthFacebook.Scopes, " "))
	parameters.Add("redirect_uri", A.OauthFacebook.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", oauthStateFaceBook)
	Url.RawQuery = parameters.Encode()
	url := Url.String()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func CallbackFacebook(w http.ResponseWriter, r *http.Request) {
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
	token, err := A.OauthFacebook.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	vals := url.Values{}
	vals.Add("fields", "id,name,email")
	vals.Add("access_token", url.QueryEscape(token.AccessToken))
	resp, err := http.Get(fmt.Sprintf("https://graph.facebook.com/%s/me?%s", A.Config.Facebook.APIVersion, vals.Encode()))
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
	accessToken := GenerateAccessToken()
	hashAccToken := CalculateSignature(accessToken, "provider")
	//check user registration
	var u models.User
	result := u.GetUser(A.DB, map[string]interface{}{
		"login":    respMap["id"],
		"provider": "facebook",
	})
	//if user not found, register new user
	if result.Error != nil || result.RowsAffected == 0 {
		u = models.User{
			Login:       respMap["id"].(string),
			Provider:    "facebook",
			Name:        respMap["name"].(string),
			AccessToken: hashAccToken,
		}
		result = u.CreateUser(A.DB)
	} else {
		u.AccessToken = hashAccToken
		u.UpdateAccessToken(A.DB)
	}
	//write cookies
	if result.Error == nil {
		var expiration = time.Now().Add(30 * 24 * time.Hour)
		cookieUID := http.Cookie{Name: "UAAT", Value: accessToken, Expires: expiration, Path: "/"}
		http.SetCookie(w, &cookieUID)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
