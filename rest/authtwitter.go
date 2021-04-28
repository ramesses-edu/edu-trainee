package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func buildAuthHeader(method, path string, params map[string]string) string {
	vals := url.Values{}
	vals.Add("oauth_consumer_key", a.Config.Twitter.TwitterAPIKey)
	vals.Add("oauth_nonce", generateNonce())
	vals.Add("oauth_signature_method", "HMAC-SHA1")
	vals.Add("oauth_timestamp", strconv.Itoa(int(time.Now().Unix())))
	vals.Add("oauth_token", a.Config.Twitter.TwitterTokenKey)
	vals.Add("oauth_version", "1.0")
	for k, v := range params {
		vals.Set(k, v)
	}
	parameterString := strings.Replace(vals.Encode(), "+", "%20", -1)
	signatureBase := strings.ToUpper(method) + "&" + url.QueryEscape(path) + "&" + url.QueryEscape(parameterString)
	signingKey := url.QueryEscape(a.Config.Twitter.TwitterAPISecret) + "&" + url.QueryEscape(a.Config.Twitter.TwitterTokenSecret)
	signature := calculateSignature(signatureBase, signingKey)
	vals.Add("oauth_signature", signature)
	returnString := "OAuth"
	for k := range vals {
		returnString += fmt.Sprintf(` %s="%s",`, k, url.QueryEscape(vals.Get(k)))
	}
	return strings.TrimRight(returnString, ",")
}

func generateNonce() string {
	const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 48)
	for i := range b {
		b[i] = allowed[rand.Intn(len(allowed))]
	}
	return string(b)
}
func authTwitter(w http.ResponseWriter, r *http.Request) {
	reqTokUrl := a.Config.Twitter.ReqTokenURL
	request, err := http.NewRequest(http.MethodPost, reqTokUrl, nil)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	autorize := buildAuthHeader(http.MethodPost, reqTokUrl, map[string]string{"oauth_callback": a.Config.Twitter.RedirectURL})
	request.Header.Set("Authorization", autorize)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if resp.StatusCode != 200 {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	dataMap := make(map[string]string)
	data := strings.Split(string(respBody), "&")
	for _, v := range data {
		datav := strings.Split(v, "=")
		dataMap[datav[0]] = datav[1]
	}
	stateToken := dataMap["oauth_token"]
	cookie := http.Cookie{Name: "oauthstate", Value: stateToken, Expires: time.Now().Add(5 * time.Minute)}
	http.SetCookie(w, &cookie)
	/////////////////////////////////////////////////////////
	urlAuthUser := a.Config.Twitter.AuthURL + "=" + stateToken
	http.Redirect(w, r, urlAuthUser, http.StatusFound)
}

func callbackTwitter(w http.ResponseWriter, r *http.Request) {
	o_token := r.FormValue("oauth_token")
	o_verifier := r.FormValue("oauth_verifier")
	oauthstate, err := r.Cookie("oauthstate")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if o_token != (oauthstate.Value) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	reqTokUrl := a.Config.Twitter.TokenURL
	request, err := http.NewRequest(http.MethodPost, reqTokUrl, bytes.NewBuffer([]byte(fmt.Sprintf("oauth_verifier=%s", o_verifier))))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	autorize := buildAuthHeader(http.MethodPost, reqTokUrl,
		map[string]string{"oauth_token": o_token, "oauth_verifier": o_verifier})
	request.Header.Set("Authorization", autorize)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if resp.StatusCode != 200 {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	dataMap := make(map[string]string)
	data := strings.Split(string(respBody), "&")
	for _, v := range data {
		datav := strings.Split(v, "=")
		dataMap[datav[0]] = datav[1]
	}

	reqTokUrl = "https://api.twitter.com/1.1/account/verify_credentials.json"
	request, err = http.NewRequest(http.MethodGet, reqTokUrl, nil)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	autorize = buildAuthHeader(http.MethodGet, reqTokUrl,
		map[string]string{"oauth_token": dataMap["oauth_token"]})
	request.Header.Set("Authorization", autorize)
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if resp.StatusCode != 200 {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	//decode answer JSON to map
	var respMap map[string]interface{} = make(map[string]interface{})
	err = json.Unmarshal(respBody, &respMap)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	//check request error
	if _, ok := respMap["errors"]; ok {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	//generate new accessToken for user
	accessToken := generateAccessToken()
	hashAccToken := calculateSignature(accessToken, "provider")
	//check user registration
	var u user
	result := u.getUser(a.DB, map[string]interface{}{
		"login":    respMap["id_str"].(string),
		"provider": "twitter",
	})
	//if user not found, register new user
	if result.Error != nil || result.RowsAffected == 0 {
		u = user{
			Login:       respMap["id_str"].(string),
			Provider:    "twitter",
			Name:        respMap["name"].(string),
			AccessToken: hashAccToken,
		}
		result = u.createUser(a.DB)
	} else {
		u.AccessToken = hashAccToken
		u.updateAccessToken(a.DB)
	}
	//write cookies
	if result.Error == nil {
		var expiration = time.Now().Add(30 * 24 * time.Hour)
		cookieUID := http.Cookie{Name: "UAAT", Value: accessToken, Expires: expiration, Path: "/"}
		http.SetCookie(w, &cookieUID)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
