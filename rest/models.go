package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"

	"gorm.io/gorm"
)

///////////////////////////////////////////////////////////////////////////////////
type user struct {
	XMLName     xml.Name  `xml:"user" json:"-" gorm:"-"`
	ID          int       `json:"id" xml:"id" gorm:"column:id;primaryKey"`
	Login       string    `json:"login" xml:"login" gorm:"column:login;unique"`
	Provider    string    `json:"-" xml:"-" gorm:"column:provider"`
	Name        string    `json:"name" xml:"name" gorm:"column:name"`
	AccessToken string    `json:"-" xml:"-" gorm:"column:access_token"`
	Posts       []post    `xml:"-" json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Comments    []comment `xml:"-" json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (u *user) getUser(db *gorm.DB, params map[string]interface{}) *gorm.DB {
	return db.Where(params).First(&u)
}
func (u *user) createUser(db *gorm.DB) *gorm.DB {
	return db.Select("Login", "Provider", "Name", "AccessToken").Create(&u)
}
func (u *user) updateAccessToken(db *gorm.DB) *gorm.DB {
	return db.Model(&u).Updates(user{AccessToken: u.AccessToken})
}

//////////////////////////////////////////////////////////////////////////////////////
type posts struct {
	XMLName xml.Name `xml:"posts" json:"-" gorm:"-"`
	Posts   []post   `xml:"post"`
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
	return db.Select("UserID", "Title", "Body").Create(&p)
}
func (p *post) updatePost(db *gorm.DB) *gorm.DB {
	return db.Model(&p).Updates(post{UserID: p.UserID, Title: p.Title, Body: p.Body})
}
func (p *post) deletePost(db *gorm.DB) *gorm.DB {
	return db.Where("userId = ?", p.UserID).Delete(&p)
}

//////////////////////////////////////////////////////////////////////////////////////////
type comments struct {
	XMLName  xml.Name  `xml:"comments" json:"-" gorm:"-"`
	Comments []comment `xml:"comment"`
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
	PostID int    `json:"postId" gorm:"column:postId"`
	UserID int    `json:"userId" gorm:"column:userId"`
	ID     int    `json:"id" gorm:"column:id;primaryKey"`
	Name   string `json:"name" gorm:"column:name;type:VARCHAR(256)"`
	Email  string `json:"email" gorm:"column:email;type:VARCHAR(256)"`
	Body   string `json:"body" gorm:"column:body;type:VARCHAR(256)"`
}

func (c *comment) getComment(db *gorm.DB, param map[string]interface{}) *gorm.DB {
	return db.Where(param).First(&c)
}

func (c *comment) createComment(db *gorm.DB) *gorm.DB {
	reEmail := regexp.MustCompile(`^[^@]+@[^@]+\.\w{1,5}$`)
	if c.Email != "" && !reEmail.Match([]byte(c.Email)) {
		return &gorm.DB{Error: gorm.ErrInvalidValue}
	}
	if c.PostID == 0 {
		return db.Select("UserID", "Name", "Email", "Body").Create(&c)
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
	return db.Where("userId = ?", c.UserID).Delete(&c)
}
