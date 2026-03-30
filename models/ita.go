package model

import (
	"gorm.io/gorm"
)

type ITA struct {
	ID      uint   `gorm:"primarykey" json:"id"`
	Year    string `json:"year"`
	Title   string `json:"title"`
	FileURL string `json:"file_url"`

	gorm.Model `json:"-"`
}
