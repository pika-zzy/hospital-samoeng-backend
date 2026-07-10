package model

import (
	"gorm.io/gorm"
)

type Personnel struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Prefix   string `json:"prefix"`
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Role     string `json:"role"`     // ฝ่ายงาน/แผนก (ใช้ group หน้า public)
	Position string `json:"position"` // ตำแหน่ง/ชื่อโรค (แยกจากแผนก)
	ImgURL   string `json:"img_url"`

	gorm.Model `json:"-"`
}
