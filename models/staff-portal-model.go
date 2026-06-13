package model

type Staffservices struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Link 		string `json:"link"`
	Color 		string `json:"color"`
	Icon 		string `json:"icon"`
}