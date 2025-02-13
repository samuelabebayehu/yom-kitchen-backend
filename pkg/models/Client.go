package models

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

type Client struct {
	gorm.Model
	Name     string `json:"name" gorm:"not null"`
	Passcode string `json:"passcode" gorm:"unique;not null;size:4"`
	Email    string `json:"email,omitempty" gorm:"unique"`
	Phone    string `json:"phone,omitempty"`
	Address  string `json:"address,omitempty"`
	IsActive bool   `json:"is_active"`
	IsAdmin  bool   `json:"is_admin"`
}

func (c *Client) BeforeCreate(tx *gorm.DB) (err error) {
	rand.NewSource(time.Now().UnixNano())

	for {
		passcode := generatePasscode()
		var existingClient Client
		if err := tx.Where("passcode = ?", passcode).First(&existingClient).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Passcode = passcode
				break
			} else {
				return err
			}
		}
	}
	return nil
}

func generatePasscode() string {
	return fmt.Sprintf("%04d", rand.Intn(10000))
}
