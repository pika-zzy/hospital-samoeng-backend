package model

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	IsActive bool   `gorm:"default:true" json:"is_active"` // สถานะ เปิด/ปิด การใช้งาน

	// TokenVersion — ใช้เพิกถอน JWT ที่ปล่อยไปแล้ว (CRIT-01)
	// ค่านี้ถูกฝังลง JWT claims ตอน login; middleware เทียบ claim กับค่าใน DB ทุก request
	// การ bump ค่านี้ (++) = เตะ token เก่าทั้งหมดของ user คนนั้นทันที (logout ทุกอุปกรณ์)
	// bump เมื่อ: ปิดใช้งานบัญชี, logout, (อนาคต) เปลี่ยนรหัสผ่าน
	TokenVersion int `gorm:"not null;default:0" json:"-"`
}
