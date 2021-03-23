package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type App struct {
	db *gorm.DB
}

func (a *App) initApp(userDB, passDB, hostDB, portDB, nameDB string) {
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
}

///////////////////////////////////////////////////////////////////////////////////
type user struct {
	XMLName  xml.Name  `xml:"user" json:"-" gorm:"-"`
	ID       int       `json:"id" gorm:"column:id;primaryKey"`
	Login    string    `json:"login" gorm:"column:login;unique"`
	Name     string    `jsin:"name" gorm:"column:name"`
	Posts    []post    `xml:"-" json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Comments []comment `xml:"-" json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

//////////////////////////////////////////////////////////////////////////////////////
type posts struct {
	XMLName xml.Name `xml:"posts" json:"-" gorm:"-"`
	Posts   []post
}

func (pp *posts) listPosts(db *gorm.DB, param map[string]interface{}) *gorm.DB {
	return db.Where(param).Find(&pp.Posts)
}

func (pp *posts) responseJSON(w http.ResponseWriter, r *http.Request) {
	jsonB, err := json.MarshalIndent(pp.Posts, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonB))
}
func (pp *posts) responseXML(w http.ResponseWriter, r *http.Request) {
	xmlB, err := xml.MarshalIndent(pp, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(xmlB))
}

/////////////////////////////////////////////////////////////////////////////////////////
type post struct {
	XMLName  xml.Name  `xml:"post" json:"-" gorm:"-"`
	UserID   int       `json:"userId" gorm:"column:userId"`
	ID       int       `json:"id" gorm:"column:id;primaryKey"`
	Title    string    `json:"title" gorm:"column:title;type:VARCHAR(256)"`
	Body     string    `json:"body" gorm:"column:body;type:VARCHAR(256)"`
	Comments []comment `xml:"-" json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (p *post) getPost(db *gorm.DB, param map[string]interface{}) *gorm.DB {
	return db.Where(param).First(&p)
}

func (p *post) createPost(db *gorm.DB) *gorm.DB {
	if p.UserID == 0 {
		return db.Select("Title", "Body").Create(&p)
	}
	return db.Select("UserID", "Title", "Body").Create(&p)
}
func (p *post) updatePost(db *gorm.DB) *gorm.DB {
	return db.Model(&p).Updates(post{UserID: p.UserID, Title: p.Title, Body: p.Body})
}
func (p *post) deletePost(db *gorm.DB) *gorm.DB {
	return db.Delete(&p)
}

//////////////////////////////////////////////////////////////////////////////////////////
type comments struct {
	XMLName  xml.Name `xml:"comments" json:"-" gorm:"-"`
	Comments []comment
}

func (cc *comments) listComments(db *gorm.DB, param map[string]interface{}) *gorm.DB {
	return db.Where(param).Find(&cc.Comments)
}
func (cc *comments) responseJSON(w http.ResponseWriter, r *http.Request) {
	jsonB, err := json.MarshalIndent(cc.Comments, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonB))
}
func (cc *comments) responseXML(w http.ResponseWriter, r *http.Request) {

	xmlB, err := xml.MarshalIndent(cc, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(xmlB))
}

////////////////////////////////////////////////////////////////////////////////////////////////
type comment struct {
	XMLName xml.Name `xml:"comment" json:"-" gorm:"-"`
	PostID  int      `json:"postId" gorm:"column:postId"`
	UserID  int      `json:"userId" gorm:"column:userId"`
	ID      int      `json:"id" gorm:"column:id;primaryKey"`
	Name    string   `json:"name" gorm:"column:name;type:VARCHAR(256)"`
	Email   string   `json:"email" gorm:"column:email;type:VARCHAR(256)"`
	Body    string   `json:"body" gorm:"column:body;type:VARCHAR(256)"`
}

func (c *comment) getComment(db *gorm.DB, param map[string]interface{}) *gorm.DB {
	return db.Where(param).First(&c)
}

func (c *comment) createComment(db *gorm.DB) *gorm.DB {
	reEmail := regexp.MustCompile(`^[^@]+@[^@]+\.\w{1,5}$`)
	if c.Email != "" && !reEmail.Match([]byte(c.Email)) {
		return &gorm.DB{Error: gorm.ErrInvalidValue}
	}
	if c.PostID == 0 && c.UserID == 0 {
		return db.Select("Name", "Email", "Body").Create(&c)
	}
	if c.PostID == 0 {
		db.Select("UserID", "Name", "Email", "Body").Create(&c)
	}
	if c.UserID == 0 {
		db.Select("PostID", "Name", "Email", "Body").Create(&c)
	}
	return db.Select("PostID", "UserID", "Name", "Email", "Body").Create(&c)
}
func (c *comment) updateComment(db *gorm.DB) *gorm.DB {
	reEmail := regexp.MustCompile(`^[^@]+@[^@]+\.\w{1,5}$`)
	if c.Email != "" && !reEmail.Match([]byte(c.Email)) {
		return &gorm.DB{Error: gorm.ErrInvalidValue}
	}
	return db.Model(&c).Updates(comment{PostID: c.PostID, UserID: c.UserID, Name: c.Name, Email: c.Email, Body: c.Body})
}
func (c *comment) deleteComment(db *gorm.DB) *gorm.DB {
	return db.Delete(&c)
}
