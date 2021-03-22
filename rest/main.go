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

// @title Blueprint Swagger API
// @version 1.0
// @description Swagger API for Golang Project Blueprint.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email martin7.heinz@gmail.com

// @license.name MIT
// @license.url https://github.com/MartinHeinz/go-project-blueprint/blob/master/LICENSE

// @BasePath /
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
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	router.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("swagger.json"),
		httpSwagger.DeepLinking(true),
	))

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
