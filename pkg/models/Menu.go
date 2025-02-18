package models

import (
	"gorm.io/gorm"
)

type MenuItem struct {
	gorm.Model
	Name      string  `form:"name" json:"name" gorm:"unique;not null"`
	Desc      string  `form:"desc" json:"desc"`
	ImageUrl  string  `json:"image_url"`
	Price     float64 `form:"price" json:"price" gorm:"not null"`
	Category  string  `form:"category" json:"category"`
	Available bool    `form:"available" json:"available" gorm:"default:true"`
}
