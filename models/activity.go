package model

import (
	"gorm.io/gorm"
)

// MaxActivityImages — จำนวนรูปสูงสุดต่อกิจกรรม **นับรวมรูปปก (Activity.ImgURL) แล้ว**
// ดังนั้นแถวใน activity_images ได้มากสุด MaxActivityImages-1 = 11 รูป
// (12 รูปลงกริด 6x2 พอดีบนเดสก์ท็อป / 4x3 บนมือถือ)
//
// ค่านี้เป็นแหล่งเดียวของเพดาน — แก้ที่นี่แล้วต้องตามไปแก้ MAX_ACTIVITY_IMAGES
// ใน frontend (interface/activity_info.ts) กับ MAX_PER_ACTIVITY ใน
// _migration/import_activity_images.py ให้ตรงกันด้วย (มีแค่ 3 ที่)
const MaxActivityImages = 12

// โครงสร้างข้อมูลกิจกรรม (Model)
type Activity struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	// ImgURL = รูปปก ใช้โชว์ในการ์ดหน้ารวม/หน้าแรก — ยังเป็นแหล่งเดียวของรูปปกเหมือนเดิม
	ImgURL string `json:"img_url"`

	// Images = รูปเพิ่มเติม **ไม่รวมรูปปก** เรียงตาม SortOrder
	// หน้าเว็บแสดงเป็น [ImgURL, Images...] จึงไม่มีรูปปกซ้ำในตารางนี้
	Images []ActivityImage `gorm:"foreignKey:ActivityID;constraint:OnDelete:CASCADE" json:"images"`

	gorm.Model `json:"-"` // ซ่อนฟิลด์นี้เวลาแปลงเป็น JSON
}

// ActivityImage — รูปเพิ่มเติมของกิจกรรม (อัลบั้ม)
//
// เดิมกิจกรรมเก็บรูปได้ใบเดียว ตอนย้ายข้อมูลจากเว็บเก่าจึงเหลือรูปแรกของอัลบั้ม
// ตารางนี้รับรูปที่เหลือ โดยไม่แตะ Activity.ImgURL — ของเดิมไม่พัง
//
// NOTE: column ชื่อ "sort_order" ไม่ใช่ "order" เพราะ ORDER เป็น reserved keyword ของ SQL
// (กติกาเดียวกับ HeroSlide)
type ActivityImage struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	ActivityID uint   `gorm:"index;not null" json:"activity_id"`
	ImgURL     string `gorm:"not null;default:''" json:"img_url"`
	SortOrder  int    `gorm:"column:sort_order;not null;default:0" json:"sort_order"`

	gorm.Model `json:"-"`
}
