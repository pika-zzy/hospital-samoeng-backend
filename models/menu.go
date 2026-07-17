package model

import "time"

type Menu struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	MenuName    string    `json:"menu_name" gorm:"column:menu_name;not null"`
	Description string    `json:"description" gorm:"column:description"`
	Link        string    `json:"link" gorm:"column:link"`
	Icon        string    `json:"icon" gorm:"column:icon"`
	Color       string    `json:"color" gorm:"column:color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Menu) TableName() string {
	return "menu"
}
