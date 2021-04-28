package models

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"

	"gorm.io/gorm"
)

type User struct {
	XMLName     xml.Name  `xml:"user" json:"-" gorm:"-"`
	ID          int       `json:"id" xml:"id" gorm:"column:id;primaryKey"`
	Login       string    `json:"login" xml:"login" gorm:"column:login;unique"`
	Provider    string    `json:"-" xml:"-" gorm:"column:provider"`
	Name        string    `json:"name" xml:"name" gorm:"column:name"`
	AccessToken string    `json:"-" xml:"-" gorm:"column:access_token"`
	Posts       []Post    `xml:"-" json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Comments    []Comment `xml:"-" json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (u *User) GetUser(db *gorm.DB, params map[string]interface{}) *gorm.DB {
	return db.Where(params).First(&u)
}
func (u *User) CreateUser(db *gorm.DB) *gorm.DB {
	return db.Select("Login", "Provider", "Name", "AccessToken").Create(&u)
}
func (u *User) UpdateAccessToken(db *gorm.DB) *gorm.DB {
	return db.Model(&u).Updates(User{AccessToken: u.AccessToken})
}

//////////////////////////////////////////////////////////////////////////////////////
type Posts struct {
	XMLName xml.Name `xml:"posts" json:"-" gorm:"-"`
	Posts   []Post   `xml:"post"`
}

func (pp *Posts) ListPosts(db *gorm.DB, param map[string]interface{}) *gorm.DB {
	return db.Where(param).Find(&pp.Posts)
}

func (pp *Posts) ResponseJSON(w http.ResponseWriter, r *http.Request) {
	jsonB, err := json.MarshalIndent(pp.Posts, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonB))
}
func (pp *Posts) ResponseXML(w http.ResponseWriter, r *http.Request) {
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
type Post struct {
	UserID   int       `json:"userId" gorm:"column:userId"`
	ID       int       `json:"id" gorm:"column:id;primaryKey"`
	Title    string    `json:"title" gorm:"column:title;type:VARCHAR(256)"`
	Body     string    `json:"body" gorm:"column:body;type:VARCHAR(256)"`
	Comments []Comment `xml:"-" json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (p *Post) GetPost(db *gorm.DB, param map[string]interface{}) *gorm.DB {
	return db.Where(param).First(&p)
}

func (p *Post) CreatePost(db *gorm.DB) *gorm.DB {
	return db.Select("UserID", "Title", "Body").Create(&p)
}
func (p *Post) UpdatePost(db *gorm.DB) *gorm.DB {
	return db.Model(&p).Updates(Post{UserID: p.UserID, Title: p.Title, Body: p.Body})
}
func (p *Post) DeletePost(db *gorm.DB) *gorm.DB {
	return db.Where("userId = ?", p.UserID).Delete(&p)
}

//////////////////////////////////////////////////////////////////////////////////////////
type Comments struct {
	XMLName  xml.Name  `xml:"comments" json:"-" gorm:"-"`
	Comments []Comment `xml:"comment"`
}

func (cc *Comments) ListComments(db *gorm.DB, param map[string]interface{}) *gorm.DB {
	return db.Where(param).Find(&cc.Comments)
}
func (cc *Comments) ResponseJSON(w http.ResponseWriter, r *http.Request) {
	jsonB, err := json.MarshalIndent(cc.Comments, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonB))
}
func (cc *Comments) ResponseXML(w http.ResponseWriter, r *http.Request) {

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
type Comment struct {
	PostID int    `json:"postId" gorm:"column:postId"`
	UserID int    `json:"userId" gorm:"column:userId"`
	ID     int    `json:"id" gorm:"column:id;primaryKey"`
	Name   string `json:"name" gorm:"column:name;type:VARCHAR(256)"`
	Email  string `json:"email" gorm:"column:email;type:VARCHAR(256)"`
	Body   string `json:"body" gorm:"column:body;type:VARCHAR(256)"`
}

func (c *Comment) GetComment(db *gorm.DB, param map[string]interface{}) *gorm.DB {
	return db.Where(param).First(&c)
}

func (c *Comment) CreateComment(db *gorm.DB) *gorm.DB {
	reEmail := regexp.MustCompile(`^[^@]+@[^@]+\.\w{1,5}$`)
	if c.Email != "" && !reEmail.Match([]byte(c.Email)) {
		return &gorm.DB{Error: gorm.ErrInvalidValue}
	}
	if c.PostID == 0 {
		return db.Select("UserID", "Name", "Email", "Body").Create(&c)
	}
	return db.Select("PostID", "UserID", "Name", "Email", "Body").Create(&c)
}
func (c *Comment) UpdateComment(db *gorm.DB) *gorm.DB {
	reEmail := regexp.MustCompile(`^[^@]+@[^@]+\.\w{1,5}$`)
	if c.Email != "" && !reEmail.Match([]byte(c.Email)) {
		return &gorm.DB{Error: gorm.ErrInvalidValue}
	}
	return db.Model(&c).Updates(Comment{PostID: c.PostID, UserID: c.UserID, Name: c.Name, Email: c.Email, Body: c.Body})
}
func (c *Comment) DeleteComment(db *gorm.DB) *gorm.DB {
	return db.Where("userId = ?", c.UserID).Delete(&c)
}
