package main

import (
	"context"
	"log"
	"net/http"
	"regexp"
)

var (
	reNum  *regexp.Regexp = regexp.MustCompile(`\d+`)
	userDB string         = "utest"
	passDB string         = "12345"
	hostDB string         = "localhost"
	portDB string         = "3306"
	nameDB string         = "edudb"
	a      App
)

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
