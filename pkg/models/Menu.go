package models

import (
	"gorm.io/gorm"
)

type MenuItem struct {
	gorm.Model
	Name      string  `json:"name" gorm:"unique;not null"`
	Desc      string  `json:"desc"`
	ImageUrl  string  `json:"image_url"`
	Price     float64 `json:"price" gorm:"not null"`
	Category  string  `json:"category"`
	Available bool    `json:"available" gorm:"default:true"`
}
