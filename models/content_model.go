package model

import (
	"gorm.io/gorm"
)

// หน้าเนื้อหาที่แก้ได้จากหลังบ้าน (Content Page)
//
// ที่มา: เว็บเก่ามี "กล่องเมนู" ข้างหน้าแรกที่ไม่ใช่ข่าว/กิจกรรม/ITA เช่น
// ความปลอดภัยด้านยา · ชมรมจริยธรรม (แยกรายปี) · PDPA รวม 20 หน้า
// ของพวกนี้เพิ่มทุกปี (ชมรมจริยธรรม 2570, DUE ปีงบใหม่) จึงต้องอยู่ใน DB
// ไม่ใช่ฝังในโค้ดแบบ src/interface/document.ts
//
// 3 ชั้น: Section (= 1 หน้าเว็บ) → Group (= 1 หัวข้อ) → File (= 1 เอกสาร)
// เหตุที่ต้องมีชั้น Group คั่น ไม่ใช่แขวนไฟล์กับ Section ตรง ๆ:
// หัวข้อเดียวมักมีหลายไฟล์ที่ต่างกันแค่ปี (DUE 2566 / DUE 2567) และบางหัวข้อ
// ไม่มีไฟล์เลยมีแต่ข้อความ (Privacy Notice) — Group เป็นที่เก็บทั้งสองแบบ

// ContentSection — 1 แถว = 1 หน้าเว็บ เสิร์ฟที่ /about/<Slug>
type ContentSection struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// Slug = ส่วนท้าย URL ของหน้า (ascii, ไม่ซ้ำ) — frontend ใช้ route /about/$slug
	// ห้ามแก้หลังเผยแพร่ ลิงก์ที่คนบันทึกไว้จะพัง
	Slug        string `gorm:"size:64;uniqueIndex;not null" json:"slug"`
	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description"`
	SortOrder   int    `gorm:"column:sort_order;not null;default:0" json:"sort_order"`

	Groups []ContentGroup `gorm:"foreignKey:SectionID;constraint:OnDelete:CASCADE" json:"groups"`

	gorm.Model `json:"-"`
}

// ContentGroup — 1 หัวข้อในหน้า (เท่ากับ 1 หน้าของเว็บเก่า)
type ContentGroup struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	SectionID uint `gorm:"index;not null" json:"section_id"`

	Title string `gorm:"not null" json:"title"`

	// Body — เนื้อความยาวที่ไม่ได้อยู่ในไฟล์แนบ (เช่น คำประกาศความเป็นส่วนตัว
	// ของเว็บเก่ายาว ~14,800 ตัวอักษร ไม่มี PDF เลย) ปกติเว้นว่าง
	//
	// ใช้ longtext ไม่ใช่ text: TEXT เก็บได้ 65,535 **ไบต์** ไม่ใช่ตัวอักษร
	// ภาษาไทยใน utf8mb4 กินตัวละ 3 ไบต์ → เหลือจริงราว 21,000 ตัวอักษร
	// ของที่มีอยู่ตอนนี้ 14,800 ตัวใกล้เพดานเกินไปสำหรับข้อความที่คนแก้ได้เอง
	Body string `gorm:"type:longtext" json:"body"`

	// Year = พ.ศ. — nil แปลว่าหัวข้อนี้ไม่ผูกกับปี (เช่นความปลอดภัยด้านยา)
	// ถ้าหัวข้อในหน้าเดียวกันมีปี frontend จะขึ้นปุ่มสลับปีให้เอง
	// ทำให้ปี 2570 เพิ่มได้จากหลังบ้านโดยไม่ต้องแตะเมนูหรือโค้ด
	Year *int `gorm:"index" json:"year"`

	SortOrder int `gorm:"column:sort_order;not null;default:0" json:"sort_order"`

	Files []ContentFile `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE" json:"files"`

	gorm.Model `json:"-"`
}

// ContentFile — เอกสาร/รูป 1 ไฟล์ใต้หัวข้อ
type ContentFile struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	GroupID uint `gorm:"index;not null" json:"group_id"`

	// Label = ชื่อที่แสดงบนหน้าเว็บ เว็บเก่าเขียนลิงก์ว่า "ดาวน์โหลดไฟล์" ทุกอัน
	// ตอน import จึงเอาบรรทัดกำกับเหนือลิงก์มาใช้แทน (เช่น "DUE ปีงบประมาณ 2566")
	Label   string `gorm:"not null" json:"label"`
	FileURL string `gorm:"not null" json:"file_url"`

	SortOrder int `gorm:"column:sort_order;not null;default:0" json:"sort_order"`

	gorm.Model `json:"-"`
}
