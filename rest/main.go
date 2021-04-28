package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"edu-trainee/rest/application"
	"edu-trainee/rest/docs"

	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/oauth2"
)

var a application.Application

func init() {
	// loads values from .env into the system
	if err := godotenv.Load("./config.env"); err != nil {
		log.Print("No .env file found")
		os.Exit(1)
	}
}

// @title Education Forum API
// @version 1.0
// @description This is a education forum server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email romgrishin@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:80
// @BasePath /
// @query.collection.format multi

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name APIKey
func main() {

	a = application.Application{}
	a.InitApplication()
	oauthGoogle = &oauth2.Config{
		ClientID:     a.Config.Google.ClientID,
		ClientSecret: a.Config.Google.ClientSecret,
		RedirectURL:  a.Config.Google.RedirectURL,
		Scopes:       a.Config.Google.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  a.Config.Google.AuthURL,
			TokenURL: a.Config.Google.TokenURL,
		},
	}
	oauthFacebook = &oauth2.Config{
		ClientID:     a.Config.Facebook.ClientID,
		ClientSecret: a.Config.Facebook.ClientSecret,
		RedirectURL:  a.Config.Facebook.RedirectURL,
		Scopes:       a.Config.Facebook.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  a.Config.Facebook.AuthURL,
			TokenURL: a.Config.Facebook.TokenURL,
		},
	}
	sql, err := a.DB.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sql.Close()
	initializeRoutes(a.Router)

	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}

	//echo
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go echoStart(ctx, "localhost:8081")

	a.ListenAndServe()
}

func initializeRoutes(router *http.ServeMux) {
	router.Handle("/", mainHandler())
	router.Handle("/public", http.NotFoundHandler())
	router.Handle("/public/", publicHandler())
	router.Handle("/logout/", logoutHandler())
	router.Handle("/auth/", mwAuthentification())
	router.Handle("/posts", mwAutorization(http.HandlerFunc(postsHandler)))
	router.Handle("/posts/", mwAutorization(http.HandlerFunc(postsHandler)))
	router.Handle("/comments", mwAutorization(http.HandlerFunc(commentsHandler)))
	router.Handle("/comments/", mwAutorization(http.HandlerFunc(commentsHandler)))
	router.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("localhost/swagger/doc.json"),
	))
}
