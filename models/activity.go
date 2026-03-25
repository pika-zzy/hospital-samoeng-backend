package model

import (
	"gorm.io/gorm"
)

// โครงสร้างข้อมูลกิจกรรม (Model)
type Activity struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	ImgURL      string `json:"img_url"`

	gorm.Model `json:"-"` // ซ่อนฟิลด์นี้เวลาแปลงเป็น JSON
}
