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
	a      Application
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

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name APIKey
func main() {
	a = Application{}
	a.initApp(userDB, passDB, hostDB, portDB, nameDB)
	sql, err := a.db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sql.Close()
	initializeRoutes(a.Router)

	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}

	// echo
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go echoStart(ctx, "localhost:8081")

	a.ListenAndServe("localhost:80")
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
