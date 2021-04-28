package application

import (
	"edu-trainee/rest/config"
	"edu-trainee/rest/models"
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
	initDBTables(a.DB)
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

func initDBTables(db *gorm.DB) {
	if !db.Migrator().HasTable(&models.Comment{}) {
		db.Migrator().CreateTable(&models.Comment{})
	}
	if !db.Migrator().HasTable(&models.Post{}) {
		db.Migrator().CreateTable(&models.Post{})
		db.Migrator().CreateConstraint(&models.Post{}, "Comments")
	} else {
		if !db.Migrator().HasConstraint(&models.Post{}, "Comments") {
			db.Migrator().CreateConstraint(&models.Post{}, "Comments")
		}
	}

	if !db.Migrator().HasTable(&models.User{}) {
		db.Migrator().CreateTable(&models.User{})
		db.Migrator().CreateConstraint(&models.User{}, "Posts")
		db.Migrator().CreateConstraint(&models.User{}, "Comments")
	} else {
		if !db.Migrator().HasConstraint(&models.User{}, "Posts") {
			db.Migrator().CreateConstraint(&models.User{}, "Posts")
		}
		if !db.Migrator().HasConstraint(&models.User{}, "Comments") {
			db.Migrator().CreateConstraint(&models.User{}, "Comments")
		}
	}
}
