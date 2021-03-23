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

// @host 127.0.0.1:80
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
	router.HandleFunc("/", mainHandler)
	router.HandleFunc("/posts/", postsHandler)
	router.HandleFunc("/comments/", commentsHandler)

	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}
	router.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("127.0.0.1/swagger/doc.json"),
	))
	//
	//router.Handle("/login", middleware(http.HandlerFunc(postsHandler))) // may midw1(midw2(final))

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

// func middleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// logic code or return
// 		next.ServeHTTP(w, r)
// 		// logic code or return
// 	})
// }
