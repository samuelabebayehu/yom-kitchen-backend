package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	ClientID    int         `json:"client_id" gorm:"not null"`
	Client      Client      `json:"client" gorm:"foreignKey:ClientID;references:ID"`
	OrderDate   time.Time   `json:"order_date" gorm:"not null;default:now()"`
	OrderItems  []OrderItem `json:"order_items" gorm:"foreignKey:OrderID;references:ID;constraint:OnDelete:CASCADE"`
	TotalAmount float64     `json:"total_amount" gorm:"not null;type:decimal(10,2)"`
	Status      string      `json:"status" gorm:"default:'Pending'"`
	Notes       string      `json:"notes,omitempty"`
}

type OrderItem struct {
	gorm.Model
	OrderID    int     `json:"order_id" gorm:"not null"`
	Order      Order   `json:"order" gorm:"foreignKey:OrderID;references:ID"`
	MenuItemID int     `json:"menu_item_id" gorm:"not null"`
	ItemName   string  `json:"item_name" gorm:"not null"`
	ItemPrice  float64 `json:"item_price" gorm:"not null;type:decimal(10,2);"`
	Quantity   int     `json:"quantity" gorm:"not null;default:1"`
	Subtotal   float64 `json:"subtotal" gorm:"not null;type:decimal(10,2);"`
}
