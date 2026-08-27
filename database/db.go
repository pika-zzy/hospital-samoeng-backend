package database

import (
	"fmt"
	model "hospitalbackend/models"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB // ตัวแปร Global เอาไว้เรียกใช้ Database จากที่อื่น

func ConnectDB() {
	// ข้อมูลเชื่อมต่อ (DSN)
	// user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// ถ้าตั้งรหัสผ่าน MySQL ไว้ ให้ใส่ตรง root:password (ปกติ XAMPP จะไม่มีรหัส ก็ใส่ว่างไว้แบบนี้ root:)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("เชื่อมต่อ Database ไม่สำเร็จ: ", err)
	}

	fmt.Println("เชื่อมต่อ Database สำเร็จแล้ว! 🎉")

	// Auto Migrate จะทำงานเฉพาะเมื่อ env RUN_MIGRATE=true เท่านั้น
	// ปกติ (บน production) ปล่อยว่าง/false → ไม่แตะ schema เลย
	// เวลาจะเพิ่ม/แก้ตาราง: ตั้ง RUN_MIGRATE=true ชั่วคราว → deploy รอบเดียว → ตั้งกลับเป็น false
	if os.Getenv("RUN_MIGRATE") == "true" {
		runAutoMigrate()
	} else {
		fmt.Println("ข้าม AutoMigrate (RUN_MIGRATE != true) — ใช้ schema เดิมของ Database")
	}
}

// runAutoMigrate: สร้าง/อัปเดตตารางตาม Struct ที่เราเขียน (เฉพาะเพิ่ม column/index ที่ขาด ไม่ลบข้อมูล)
// เพิ่ม model ใหม่ ให้มาเติมในวงเล็บนี้
func runAutoMigrate() {
	// FIX: เดิมไม่เช็ค error ของ AutoMigrate — ถ้า migrate ตารางไหนพัง จะเงียบ (ตารางไม่ถูกสร้างโดยไม่มีสัญญาณ)
	if err := DB.AutoMigrate(
		&model.ITAYear{},
		&model.News{},
		&model.Activity{},
		&model.ActivityImage{},
		&model.Personnel{},
		&model.User{},
		&model.MoitCategory{},
		&model.MoitTopic{},
		&model.MoitItem{},
		&model.ITA{},
		&model.RolePermission{},
		&model.Popup{},
		&model.HeroSlide{},
		&model.Menu{},
		&model.ContentSection{},
		&model.ContentGroup{},
		&model.ContentFile{},
	); err != nil {
		log.Fatal("AutoMigrate ไม่สำเร็จ: ", err)
	}
	DB.Exec("SET FOREIGN_KEY_CHECKS = 1")
	fmt.Println("AutoMigrate สำเร็จ ✅")
}

func SeedData() {
	var count int64
	DB.Model(&model.News{}).Count(&count)

}
