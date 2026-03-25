package model

import (
	"gorm.io/gorm"
)

type Personnel struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Prefix   string `json:"prefix"`
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Uid      int    `json:"uid"`
	Role     string `json:"role"`
	ImgURL   string `json:"img_url"`

	gorm.Model `json:"-"`
}
