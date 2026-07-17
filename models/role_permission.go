package model

// RolePermission — เมนูหลังบ้านที่ role มองเห็นได้ (1 แถวต่อ 1 เมนูต่อ role)
// role เป็น string ตรงกับ User.Role ("admin"/"employee") — admin เห็นทุกเมนูเสมอ
// ฝั่ง frontend จึงใช้ตารางนี้กับ employee เท่านั้น
// menu_key คือ NAV.to ของ sidebar ฝั่ง frontend เช่น "/admin/news/news/summary"
type RolePermission struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Role    string `gorm:"index;size:50" json:"role"`
	MenuKey string `gorm:"size:255" json:"menu_key"`
}
