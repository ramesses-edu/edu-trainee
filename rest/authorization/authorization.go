package authorization

import (
	"crypto/hmac"
	"crypto/sha1"
	"edu-trainee/rest/application"
	"edu-trainee/rest/models"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	A application.Application
)

func generateOauthStateProvider() string {
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	return state
}

func buildAuthHeader(method, path string, params map[string]string) string {
	vals := url.Values{}
	vals.Add("oauth_consumer_key", A.Config.Twitter.TwitterAPIKey)
	vals.Add("oauth_nonce", generateNonce())
	vals.Add("oauth_signature_method", "HMAC-SHA1")
	vals.Add("oauth_timestamp", strconv.Itoa(int(time.Now().Unix())))
	vals.Add("oauth_token", A.Config.Twitter.TwitterTokenKey)
	vals.Add("oauth_version", "1.0")
	for k, v := range params {
		vals.Set(k, v)
	}
	parameterString := strings.Replace(vals.Encode(), "+", "%20", -1)
	signatureBase := strings.ToUpper(method) + "&" + url.QueryEscape(path) + "&" + url.QueryEscape(parameterString)
	signingKey := url.QueryEscape(A.Config.Twitter.TwitterAPISecret) + "&" + url.QueryEscape(A.Config.Twitter.TwitterTokenSecret)
	signature := CalculateSignature(signatureBase, signingKey)
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

func CalculateSignature(base, key string) string {
	hash := hmac.New(sha1.New, []byte(key))
	hash.Write([]byte(base))
	signature := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature)
}

func GenerateAccessToken() string {
	b := make([]byte, 64)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	return state
}

func GetCurrentUser(db *gorm.DB, r *http.Request) models.User {
	accessToken := ""
	accessTokenCookie, err := r.Cookie("UAAT")
	if err == nil {
		accessToken = accessTokenCookie.Value
	}
	if accessToken == "" {
		accessToken = r.Header.Get("APIKey")
	}
	if accessToken == "" {
		return models.User{}
	}
	hashAccTok := CalculateSignature(accessToken, "provider")
	var u models.User
	result := u.GetUser(db, map[string]interface{}{
		"access_token": hashAccTok,
	})
	if result.Error != nil || result.RowsAffected == 0 {
		return models.User{}
	}
	return u
}
