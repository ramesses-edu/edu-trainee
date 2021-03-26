package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	twitterAPIKey      = "TmSwp1vBJfQeXWoAt5G6SqWmy"
	twitterAPISecret   = "0ORlrpHLPCFQCY3QPSa9ZMtW5p9OBscKpF36idOL2itJxvVaOd"
	twitterTokenKey    = "1374820281542438923-6KfgTG8HnLORpaHNwH9EGravoK4t4U"
	twitterTokenSecret = "nQN44yJd9nSvs5elZc578dpaZ0hZwrTgexsUg3SSZ1VNl"
)

func buildAuthHeader(method, path string, params map[string]string) string {
	vals := url.Values{}
	vals.Add("oauth_consumer_key", twitterAPIKey)
	vals.Add("oauth_nonce", generateNonce())
	vals.Add("oauth_signature_method", "HMAC-SHA1")
	vals.Add("oauth_timestamp", strconv.Itoa(int(time.Now().Unix())))
	vals.Add("oauth_token", twitterTokenKey)
	vals.Add("oauth_version", "1.0")
	for k, v := range params {
		vals.Set(k, v)
	}
	//fmt.Println("oauth_token buildHeader: " + vals.Get("oauth_token"))
	parameterString := strings.Replace(vals.Encode(), "+", "%20", -1)
	signatureBase := strings.ToUpper(method) + "&" + url.QueryEscape(path) + "&" + url.QueryEscape(parameterString)
	signingKey := url.QueryEscape(twitterAPISecret) + "&" + url.QueryEscape(twitterTokenSecret)
	signature := calculateSignature(signatureBase, signingKey)
	vals.Add("oauth_signature", signature)
	returnString := "OAuth"
	for k := range vals {
		returnString += fmt.Sprintf(` %s="%s",`, k, url.QueryEscape(vals.Get(k)))
	}
	return strings.TrimRight(returnString, ",")
}
func calculateSignature(base, key string) string {
	hash := hmac.New(sha1.New, []byte(key))
	hash.Write([]byte(base))
	signature := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature)
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
	reqTokUrl := "https://api.twitter.com/oauth/request_token"
	request, err := http.NewRequest(http.MethodPost, reqTokUrl, nil)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	autorize := buildAuthHeader(http.MethodPost, reqTokUrl, map[string]string{"oauth_callback": "http://localhost/auth/callback/twitter"})
	request.Header.Set("Authorization", autorize)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if resp.StatusCode != 200 {
		w.Write([]byte(resp.Status))
		//http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
	respToken := dataMap["oauth_token"]
	cookie := http.Cookie{Name: "oauthstate", Value: respToken, Expires: time.Now().Add(24 * time.Hour)}
	http.SetCookie(w, &cookie)
	//fmt.Println("Request token: " + respToken)
	urlAuthUser := "https://api.twitter.com/oauth/authenticate?oauth_token="
	http.Redirect(w, r, urlAuthUser+respToken, http.StatusFound)
}

func callbackTwitter(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "oauth_token=%s \n oauth_verifier=%s", r.FormValue("oauth_token"), )
	//fmt.Printf("oauth_token=%s \n oauth_verifier=%s", r.FormValue("oauth_token"), r.FormValue("oauth_verifier"))
	o_token := r.FormValue("oauth_token")
	o_verifier := r.FormValue("oauth_verifier")
	oauthstate, _ := r.Cookie("oauthstate")
	if o_token != oauthstate.Value {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//
	//	Как здесь определить валидность токена??

	reqTokUrl := "https://api.twitter.com/oauth/access_token"
	request, err := http.NewRequest(http.MethodPost, reqTokUrl, bytes.NewBuffer([]byte(fmt.Sprintf("oauth_verifier=%s", r.FormValue("oauth_verifier")))))
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
		w.Write([]byte(resp.Status))
		//http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	//fmt.Println(string(respBody))
	dataMap := make(map[string]string)
	data := strings.Split(string(respBody), "&")
	for _, v := range data {
		datav := strings.Split(v, "=")
		dataMap[datav[0]] = datav[1]
		fmt.Println(datav[0] + ": " + datav[1])
	}
	/*
		oauth_token: 1374820281542438923-sbEtNgMwgTqhvJXFD3LBL2rSNXBTAE
		oauth_token_secret: E9jp8WXrjLMkvF5vaAMOUtYD6hnPfpFUHN6mPMwK5vVf2
		user_id: 1374820281542438923
		screen_name: Roman07334929

		При ошибке: код 401
	*/
	var expiration = time.Now().Add(30 * 24 * time.Hour)
	cookieProvider := http.Cookie{Name: "provider", Value: "twitter", Expires: expiration, Path: "/"}
	cookieToken := http.Cookie{Name: "TASID", Value: dataMap["oauth_token"], Expires: expiration, Path: "/"}
	//cookieUID := http.Cookie{Name: "SAUID", Value: dataMap["user_id"], Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookieProvider)
	http.SetCookie(w, &cookieToken)
	//http.SetCookie(w, &cookieUID)

	reqTokUrl = "https://api.twitter.com/1.1/account/verify_credentials.json"
	request, err = http.NewRequest(http.MethodGet, reqTokUrl, nil)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	autorize = buildAuthHeader(http.MethodGet, reqTokUrl,
		map[string]string{})
	request.Header.Set("Authorization", autorize)
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if resp.StatusCode != 200 {
		w.Write([]byte(resp.Status))
		//http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	fmt.Println(string(respBody))
}
