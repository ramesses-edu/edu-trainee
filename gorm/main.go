package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	userDB   string = "utest"
	passDB   string = "12345"
	hostDB   string = "localhost"
	portDB   string = "3306"
	nameDB   string = "edudb"
	gormOpen        = func() *gorm.DB {
		gormDialector := mysql.New(mysql.Config{
			DSN: fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", userDB, passDB, hostDB, portDB, nameDB),
		})
		db, err := gorm.Open(gormDialector, &gorm.Config{})
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return db
	}
	jsonplaceholder string = "https://jsonplaceholder.typicode.com/"
	mux             sync.Mutex
)

type post struct {
	UserID int    `json:"userId" gorm:"column:userId"`
	ID     int    `json:"id" gorm:"column:id;primaryKey"`
	Title  string `json:"title" gorm:"column:title;type:VARCHAR(256)"`
	Body   string `json:"body" gorm:"column:body;type:VARCHAR(256)"`
}
type comment struct {
	PostID int    `json:"postId" gorm:"column:postId"`
	ID     int    `json:"id" gorm:"column:id;primaryKey"`
	Name   string `json:"name" gorm:"column:name;type:VARCHAR(256)"`
	Email  string `json:"email" gorm:"column:email;type:VARCHAR(256)"`
	Body   string `json:"body" gorm:"column:body;type:VARCHAR(256)"`
}

func main() {
	postsURI := "posts?userId=7"
	url := jsonplaceholder + postsURI
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	var posts []post
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(respBody, &posts)
	if err != nil {
		fmt.Println(err)
	}
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	db := gormOpen()
	if db == nil {
		return
	}
	sql, _ := db.DB()
	defer sql.Close()
	db.Migrator().DropTable(&post{})
	db.Migrator().DropTable(&comment{})
	db.Migrator().CreateTable(&post{})
	db.Migrator().CreateTable(&comment{})
	for _, value := range posts {
		wg.Add(1)
		go func(ctx context.Context, p post, db *gorm.DB) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				mux.Lock()
				db.Create(&p)
				mux.Unlock()
				commentURI := "comments?postId="
				url = jsonplaceholder + commentURI + strconv.Itoa(p.ID)
				var comments []comment
				resp, err := http.Get(url)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer resp.Body.Close()
				respBody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = json.Unmarshal(respBody, &comments)
				if err != nil {
					fmt.Println(err)
				}
				var wg2 sync.WaitGroup
				ctx2, cancel := context.WithTimeout(ctx, time.Second*5)
				defer cancel()
				for _, value := range comments {
					wg2.Add(1)
					go func(ctx context.Context, c comment, db *gorm.DB) {
						defer wg2.Done()
						select {
						case <-ctx.Done():
							return
						default:
							mux.Lock()
							db.Create(&c)
							mux.Unlock()
						}
					}(ctx2, value, db)
				}
				wg2.Wait()
			}
		}(ctx, value, db)
	}
	wg.Wait()
}
