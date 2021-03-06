package middleware

import (
	"edu-trainee/rest/application"
	"edu-trainee/rest/authorization"
	"edu-trainee/rest/models"
	"net/http"
)

func Autorization(next http.Handler) http.Handler {
	db := application.CurrentApplication().DB
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
			w.WriteHeader(http.StatusNetworkAuthenticationRequired)
			w.Write([]byte(`{"error":""}`))
			return
		}
		hashAccTok := authorization.CalculateSignature(accessToken, "provider")
		var u models.User
		result := u.GetUser(db, map[string]interface{}{
			"access_token": hashAccTok,
		})
		if result.Error != nil || result.RowsAffected == 0 {
			w.WriteHeader(http.StatusNetworkAuthenticationRequired)
			w.Write([]byte(`{"error":""}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
