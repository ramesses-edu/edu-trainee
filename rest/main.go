package main

import (
	"context"
	"log"
	"net/http"

	"edu-trainee/rest/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

var (
	userDB string = "utest"
	passDB string = "12345"
	hostDB string = "localhost"
	portDB string = "3306"
	nameDB string = "edudb"
	a      App
)

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

// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token

// @securitydefinitions.oauth2.accessCode OAuth2AccessCode
// @tokenUrl https://example.com/oauth/token
// @authorizationurl https://example.com/oauth/authorize
func main() {
	a = App{}
	a.initApp(userDB, passDB, hostDB, portDB, nameDB)
	sql, err := a.db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sql.Close()
	router := http.NewServeMux()
	router.Handle("/", http.FileServer(http.Dir("./static")))
	router.Handle("/auth/", mwAuthentification())
	//	router.Handle("/callback/", nil)
	router.Handle("/posts/", mwAutorization(mwValidateToken(http.HandlerFunc(postsHandler))))
	router.Handle("/comments/", mwAutorization(mwValidateToken(http.HandlerFunc(commentsHandler))))

	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}
	router.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("localhost/swagger/doc.json"),
	))

	// echo
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go echoStart(ctx, "localhost:8081")

	server := http.Server{
		Handler: router,
		Addr:    "localhost:80",
	}
	server.ListenAndServe()
}
