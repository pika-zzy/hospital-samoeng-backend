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
}
