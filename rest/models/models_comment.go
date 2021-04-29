package models

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"

	"gorm.io/gorm"
)

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
