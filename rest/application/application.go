package application

import (
	"edu-trainee/rest/config"
	"fmt"
	"log"
	"net/http"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Application struct {
	DB     *gorm.DB
	Router *http.ServeMux
	Server *http.Server
	Config *config.Config
}

func (a *Application) InitApplication() {
	//init configuration
	a.Config = config.New()
	//init DB connection
	gormDialector := mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", a.Config.DB.UserDB, a.Config.DB.PassDB, a.Config.DB.HostDB, a.Config.DB.PortDB, a.Config.DB.NameDB),
	})
	var err error
	a.DB, err = gorm.Open(gormDialector, &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}	
	//init DB tables
	// if !a.db.Migrator().HasTable(&comment{}) {
	// 	a.db.Migrator().CreateTable(&comment{})
	// }
	// if !a.db.Migrator().HasTable(&post{}) {
	// 	a.db.Migrator().CreateTable(&post{})
	// 	a.db.Migrator().CreateConstraint(&post{}, "Comments")
	// } else {
	// 	if !a.db.Migrator().HasConstraint(&post{}, "Comments") {
	// 		a.db.Migrator().CreateConstraint(&post{}, "Comments")
	// 	}
	// }

	// if !a.db.Migrator().HasTable(&user{}) {
	// 	a.db.Migrator().CreateTable(&user{})
	// 	a.db.Migrator().CreateConstraint(&user{}, "Posts")
	// 	a.db.Migrator().CreateConstraint(&user{}, "Comments")
	// } else {
	// 	if !a.db.Migrator().HasConstraint(&user{}, "Posts") {
	// 		a.db.Migrator().CreateConstraint(&user{}, "Posts")
	// 	}
	// 	if !a.db.Migrator().HasConstraint(&user{}, "Comments") {
	// 		a.db.Migrator().CreateConstraint(&user{}, "Comments")
	// 	}
	// }

	//init Router
	a.Router = http.NewServeMux()
}

func (a *Application) ListenAndServe() {
	a.Server = &http.Server{
		Handler: a.Router,
		Addr:    a.Config.HostAddr,
	}
	a.Server.ListenAndServe()
}
