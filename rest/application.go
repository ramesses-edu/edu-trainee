package main

import (
	"fmt"
	"log"
	"net/http"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Application struct {
	db     *gorm.DB
	Router *http.ServeMux
	Server *http.Server
}

func (a *Application) initApp(userDB, passDB, hostDB, portDB, nameDB string) {
	//init DB connection
	gormDialector := mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", userDB, passDB, hostDB, portDB, nameDB),
	})
	var err error
	a.db, err = gorm.Open(gormDialector, &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	//init DB tables
	if !a.db.Migrator().HasTable(&comment{}) {
		a.db.Migrator().CreateTable(&comment{})
	}
	if !a.db.Migrator().HasTable(&post{}) {
		a.db.Migrator().CreateTable(&post{})
		a.db.Migrator().CreateConstraint(&post{}, "Comments")
	} else {
		if !a.db.Migrator().HasConstraint(&post{}, "Comments") {
			a.db.Migrator().CreateConstraint(&post{}, "Comments")
		}
	}

	if !a.db.Migrator().HasTable(&user{}) {
		a.db.Migrator().CreateTable(&user{})
		a.db.Migrator().CreateConstraint(&user{}, "Posts")
		a.db.Migrator().CreateConstraint(&user{}, "Comments")
	} else {
		if !a.db.Migrator().HasConstraint(&user{}, "Posts") {
			a.db.Migrator().CreateConstraint(&user{}, "Posts")
		}
		if !a.db.Migrator().HasConstraint(&user{}, "Comments") {
			a.db.Migrator().CreateConstraint(&user{}, "Comments")
		}
	}
	// init Router
	a.Router = http.NewServeMux()
}
func (a *Application) ListenAndServe(addr string) {
	a.Server = &http.Server{
		Handler: a.Router,
		Addr:    addr,
	}
	a.Server.ListenAndServe()
}
