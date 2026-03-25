package model

import "gorm.io/gorm"

// โครงสร้างข้อมูลข่าว (Model)
type News struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Type        string `json:"type"`
	ImgURL      string `json:"img_url"`
	FileURL     string `json:"file_url"`

	gorm.Model `json:"-"` // ซ่อนฟิลด์นี้เวลาแปลงเป็น JSON

}
