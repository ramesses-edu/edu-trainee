package application

import (
	"edu-trainee/rest/config"
	"edu-trainee/rest/models"
	"errors"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var a *Application

type Application struct {
	DB            *gorm.DB
	Router        *http.ServeMux
	Server        *http.Server
	Config        *config.Config
	OauthGoogle   *oauth2.Config
	OauthFacebook *oauth2.Config
}

func (a *Application) initApplication() {
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
	a.OauthGoogle = &oauth2.Config{
		ClientID:     a.Config.Google.ClientID,
		ClientSecret: a.Config.Google.ClientSecret,
		RedirectURL:  a.Config.Google.RedirectURL,
		Scopes:       a.Config.Google.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  a.Config.Google.AuthURL,
			TokenURL: a.Config.Google.TokenURL,
		},
	}
	a.OauthFacebook = &oauth2.Config{
		ClientID:     a.Config.Facebook.ClientID,
		ClientSecret: a.Config.Facebook.ClientSecret,
		RedirectURL:  a.Config.Facebook.RedirectURL,
		Scopes:       a.Config.Facebook.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  a.Config.Facebook.AuthURL,
			TokenURL: a.Config.Facebook.TokenURL,
		},
	}
}

func (a *Application) ListenAndServe() {
	a.Server = &http.Server{
		Handler: a.Router,
		Addr:    a.Config.HostAddr,
	}
	a.Server.ListenAndServe()
}

func New() *Application {
	a = &Application{}
	a.initApplication()
	return a
}
func Close() error {
	if a == nil {
		return errors.New("nil application")
	}
	sql, _ := a.DB.DB()
	sql.Close()
	a = nil
	return nil
}
func CurrentApplication() *Application {
	return a
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
