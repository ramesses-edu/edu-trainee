package models

import (
	"encoding/xml"

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
