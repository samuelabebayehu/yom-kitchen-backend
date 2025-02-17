package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string `json:"username" gorm:"unique;not null"`
	PasswordHash string `json:"password" gorm:"not null"`
	IsAdmin      bool   `json:"is_admin" gorm:"default:false"`
}
