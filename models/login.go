package model

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	IsActive bool   `gorm:"default:true" json:"is_active"` // สถานะ เปิด/ปิด การใช้งาน
}
