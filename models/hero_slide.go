package model

import "gorm.io/gorm"

// HeroSlide — รูปสไลด์ hero บนหน้าแรกของเว็บ (admin จัดการเอง)
// เรียงลำดับการแสดงผลด้วย Order (น้อย → มาก)
//
// NOTE: ตั้ง column เป็น "sort_order" ไม่ใช่ "order" เพราะ ORDER เป็น reserved keyword ของ SQL
// ถ้าปล่อยให้ GORM ตั้งชื่อ column ว่า order เอง จะต้องพึ่ง backtick ตลอดและพลาดง่ายเวลาเขียน raw query
//
// กติกาสำคัญ (บังคับที่ DeleteHeroSlide): เมื่อมีรูปในระบบแล้ว ห้ามลบจนเหลือ 0 รูป
// — ต้องเหลืออย่างน้อย 1 รูปเสมอ ไม่งั้นหน้าแรกจะกลายเป็นพื้นหลังเปล่า
type HeroSlide struct {
	gorm.Model

	ImageURL string `gorm:"not null;default:''" json:"image_url"` // เช่น "/uploads/images/hero/xxx.jpg"
	Order    int    `gorm:"column:sort_order;not null;default:0" json:"order"`

	// ข้อความ overlay ต่อสไลด์ (admin แก้เอง) — ว่าง = ไม่แสดงส่วนนั้น
	// ShowText คุมทั้งก้อน: false = โชว์แค่รูปเปล่า (รูปเพิ่มใหม่ default false จนกว่าจะตั้งข้อความ)
	Badge      string `gorm:"default:''" json:"badge"`
	Title      string `gorm:"default:''" json:"title"`
	Subtitle   string `gorm:"default:''" json:"subtitle"`
	ButtonText string `gorm:"column:button_text;default:''" json:"button_text"`
	ButtonLink string `gorm:"column:button_link;default:''" json:"button_link"`
	ShowText   bool   `gorm:"column:show_text;not null;default:false" json:"show_text"`
}
