package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Application struct {
	db *gorm.DB
}

func (a *Application) initApp(userDB, passDB, hostDB, portDB, nameDB string) {
	gormDialector := mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", userDB, passDB, hostDB, portDB, nameDB),
	})
	var err error
	a.db, err = gorm.Open(gormDialector, &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	//////////////////////////////////////////////////////////////
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
	var u user = user{Login: "test", Name: "test", Provider: "test", AccessToken: calculateSignature("test", "provider")}
	if !a.db.Migrator().HasTable(&user{}) {
		a.db.Migrator().CreateTable(&user{})
		a.db.Migrator().CreateConstraint(&user{}, "Posts")
		a.db.Migrator().CreateConstraint(&user{}, "Comments")
		//a.db.Create(&u)
	} else {
		if !a.db.Migrator().HasConstraint(&user{}, "Posts") {
			a.db.Migrator().CreateConstraint(&user{}, "Posts")
		}
		if !a.db.Migrator().HasConstraint(&user{}, "Comments") {
			a.db.Migrator().CreateConstraint(&user{}, "Comments")
		}
	}
	a.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&u)
}
