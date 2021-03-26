package main

import (
	"net/http"
	"regexp"

	"gorm.io/gorm"
)

func mwAutorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		// _, err := r.Cookie("uid")
		// if err != nil {
		// 	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		// 	return
		// }
		// _, err = r.Cookie("provider")
		// if err != nil {
		// 	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		// 	return
		// }
		// _, err = r.Cookie("token")
		// if err != nil {
		// 	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		// 	return
		// }
		// next.ServeHTTP(w, r)
	})
}
func mwValidateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		// provider, _ := r.Cookie("provider")
		// token, _ := r.Cookie("token")
		// switch provider.Value {
		// case "google":
		// 	r.Form.Add(token.Value, token.Value)
		// 	// провести exchange
		// 	// запросить пользователя(при необходимости зарегить)  либо выбросить на авторизацию
		// 	// отдать на манипуляцию с данными
		// case "facebook":
		// case "twitter":
		// default:
		// 	w.WriteHeader(http.StatusNetworkAuthenticationRequired)
		// 	return
		// }
	})
}
func getCurrentUser(db *gorm.DB, uid, provider string) user {
	var u user
	result := u.getUserByLoginProvider(db, provider, uid)
	if result.Error != nil || result.RowsAffected == 0 {
		return user{}
	}
	return u
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
