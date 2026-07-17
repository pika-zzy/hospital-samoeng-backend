package model

import "gorm.io/gorm"

// Popup — ป็อปอัปประชาสัมพันธ์ของเว็บไซต์
// NOTE: ออกแบบเป็น singleton — ทั้งเว็บมีได้แค่ 1 รายการ (controller ยึด record แถวแรกเสมอ)
type Popup struct {
	gorm.Model

	Status   bool   `gorm:"not null;default:false"` // true = แสดง popup ฝั่ง public
	ImageURL string `gorm:"not null;default:''"`     // path รูปภาพ เช่น "/uploads/popup/xxx.jpg"
}
