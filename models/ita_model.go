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
type MoitItem struct {
	ID      uint   `gorm:"primaryKey"`
	Label   string `gorm:"not null"`
	TopicID uint   `gorm:"not null;index"`
	Topic MoitTopic `gorm:"constraint:OnDelete:CASCADE;"`
}

// -------- ITA FILE --------
type ITA struct {
	gorm.Model

	ItemID uint     `gorm:"not null;index"`
	Item   MoitItem `gorm:"constraint:OnDelete:CASCADE;"`

	Title   string `gorm:"not null"`
	FileURL string `gorm:"not null"`

	YearID uint    `gorm:"not null;index"`
	Year   ITAYear `gorm:"constraint:OnDelete:CASCADE;"`
}