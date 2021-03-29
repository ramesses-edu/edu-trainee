package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"gorm.io/gorm"
)

func generateOauthStateProvider() string {
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	return state
}
func generateAccessToken() string {
	b := make([]byte, 64)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	return state
}
func calculateSignature(base, key string) string {
	hash := hmac.New(sha1.New, []byte(key))
	hash.Write([]byte(base))
	signature := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature)
}
func getCurrentUser(db *gorm.DB, r *http.Request) user {
	accTokenC, err := r.Cookie("UAAT")
	if err != nil {
		return user{}
	}
	accToken := calculateSignature(accTokenC.Value, "provider")
	var u user
	result := u.getUser(db, map[string]interface{}{
		"access_token": accToken,
	})
	if result.Error != nil || result.RowsAffected == 0 {
		return user{}
	}
	return u
}

func mwAutorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := ""
		accessTokenCookie, err := r.Cookie("UAAT")
		if err == nil {
			accessToken = accessTokenCookie.Value
		}
		if accessToken == "" {
			accessToken = r.Header.Get("APIKey")
		}
		if accessToken == "" {
			cookieRedirect := http.Cookie{Name: "redirect", Value: r.URL.Path, Path: "/", Expires: time.Now().Add(5 * time.Minute)}
			http.SetCookie(w, &cookieRedirect)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		hashAccTok := calculateSignature(accessToken, "provider")
		var u user
		result := u.getUser(a.db, map[string]interface{}{
			"access_token": hashAccTok,
		})
		if result.Error != nil || result.RowsAffected == 0 {
			cookieRedirect := http.Cookie{Name: "redirect", Value: r.URL.Path, Path: "/", Expires: time.Now().Add(5 * time.Minute)}
			http.SetCookie(w, &cookieRedirect)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func mwAuthentification() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rPath := r.URL.Path
		reGoogleProvider := regexp.MustCompile(`\/auth\/google(\/)??`)
		reFacebookProvider := regexp.MustCompile(`\/auth\/facebook(\/)??`)
		reTwitterProvider := regexp.MustCompile(`\/auth\/twitter(\/)??`)
		reCallback := regexp.MustCompile(`\/auth\/callback(\/)??\w+`)
		switch {
		case reGoogleProvider.Match([]byte(rPath)):
			authGoogle(w, r)
		case reFacebookProvider.Match([]byte(rPath)):
			authFacebook(w, r)
		case reTwitterProvider.Match([]byte(rPath)):
			authTwitter(w, r)
		case reCallback.Match([]byte(rPath)):
			oauthCallback(w, r)
		}
	})
}
func oauthCallback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	reProviderGoogle := regexp.MustCompile(`\/auth\/callback\/google(\/)??`)
	reProviderFacebook := regexp.MustCompile(`\/auth\/callback\/facebook(\/)??`)
	reProviderTwitter := regexp.MustCompile(`\/auth\/callback\/twitter(\/)??`)
	switch {
	case reProviderGoogle.Match([]byte(r.URL.Path)):
		callbackGoogle(w, r)
	case reProviderFacebook.Match([]byte(r.URL.Path)):
		callbackFacebook(w, r)
	case reProviderTwitter.Match([]byte(r.URL.Path)):
		callbackTwitter(w, r)
	}
}
