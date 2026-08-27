package model

import "gorm.io/gorm"

// -------- YEAR --------
type ITAYear struct {
	ID   uint `gorm:"primaryKey"`
	Year int  `gorm:"unique;not null"`

	Moits []MoitCategory `gorm:"foreignKey:YearID;constraint:OnDelete:CASCADE"`
}

// -------- MOIT --------
type MoitCategory struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description *string
	YearID      uint   `gorm:"not null;index"`
	Year ITAYear `gorm:"constraint:OnDelete:CASCADE;"`
	Topics []MoitTopic `gorm:"foreignKey:MoitID;constraint:OnDelete:CASCADE"`
}

// -------- TOPIC --------
type MoitTopic struct {
	ID     uint   `gorm:"primaryKey"`
	Label  string `gorm:"not null"`
	MoitID uint   `gorm:"not null;index"`
	Moit MoitCategory `gorm:"constraint:OnDelete:CASCADE;"`
	Items []MoitItem `gorm:"foreignKey:TopicID;constraint:OnDelete:CASCADE"`
}

// -------- ITEM --------
// ข้อรองซ้อนกันได้ 1 ชั้น (ParentID) — รวมกับ Topic แล้วหน้าเว็บได้ 3 ชั้น:
//
//	ไตรมาสที่ 1                        <- MoitTopic
//	  1. มีบันทึกข้อความฯ   [ไฟล์]      <- MoitItem (ParentID = nil)
//	  2. มีแบบสรุปผลฯ                   <- MoitItem (ParentID = nil) ไม่มีไฟล์ของตัวเอง
//	       แสดงแบบ สขร. 1 เดือนตุลาคม   <- MoitItem (ParentID = ข้อ 2)
//	  3. มีแบบฟอร์มการเผยแพร่ฯ [ไฟล์]   <- MoitItem (ParentID = nil)
//
// เดิมโครงนี้ทำไม่ได้ ตอน scrape เลยยัดสองชั้นไว้ในชื่อ topic เดียวคั่นด้วย "›"
// ซึ่งเป็นกติกาลับที่คนกรอกหลังบ้านต้องจำเอง — ParentID มาแทนของนั้น
//
// กติกาที่ controller บังคับ (MaxItemDepth):
//   - แม่ต้องอยู่ topic เดียวกัน
//   - แม่ต้องเป็นข้อชั้นบนสุด (ParentID = nil) => ลึกได้แค่ 2 ชั้นของ item
//     กติกาข้อนี้ทำให้ "วงวน" เกิดไม่ได้เลยโดยไม่ต้องไล่ต้นไม้
//   - ข้อที่มีลูกอยู่แล้ว ห้ามย้ายไปเป็นลูกของใคร
type MoitItem struct {
	ID      uint   `gorm:"primaryKey"`
	Label   string `gorm:"not null"`
	TopicID uint   `gorm:"not null;index"`
	Topic MoitTopic `gorm:"constraint:OnDelete:CASCADE;"`

	// nil = ข้อชั้นบนสุด · มีค่า = เป็นข้อย่อยของ item ตัวนั้น
	ParentID *uint      `gorm:"index"`
	Parent   *MoitItem  `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`
	Children []MoitItem `gorm:"foreignKey:ParentID"`
}

// -------- ITA FILE --------
// ไฟล์แขวนได้ 2 แบบ — ต้องมีอย่างใดอย่างหนึ่งเท่านั้น (บังคับใน controller):
//   ItemID  = แนบกับ "ข้อรอง" เช่น 1.1, 1.2   (เคสปกติ)
//   TopicID = แนบกับ "หัวข้อ" ตรง ๆ            (เคสที่หัวข้อนั้นไม่มีข้อรอง)
// เดิม ItemID เป็น not null ทำให้ต้องสร้างข้อรองหลอกมารับไฟล์
type ITA struct {
	gorm.Model

	ItemID *uint     `gorm:"index"`
	Item   *MoitItem `gorm:"constraint:OnDelete:CASCADE;"`

	TopicID *uint      `gorm:"index"`
	Topic   *MoitTopic `gorm:"constraint:OnDelete:CASCADE;"`

	Title   string `gorm:"not null"`
	FileURL string `gorm:"not null"`

	YearID uint    `gorm:"not null;index"`
	Year   ITAYear `gorm:"constraint:OnDelete:CASCADE;"`
}