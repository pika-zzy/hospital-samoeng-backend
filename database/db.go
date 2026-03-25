package database

import (
	"fmt"
	model "hospitalbackend/models"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB // ตัวแปร Global เอาไว้เรียกใช้ Database จากที่อื่น

func ConnectDB() {
	// ข้อมูลเชื่อมต่อ (DSN)
	// user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := "root:@tcp(127.0.0.1:3306)/hospital_db?charset=utf8mb4&parseTime=True&loc=Local"

	// ถ้าตั้งรหัสผ่าน MySQL ไว้ ให้ใส่ตรง root:password (ปกติ XAMPP จะไม่มีรหัส ก็ใส่ว่างไว้แบบนี้ root:)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("เชื่อมต่อ Database ไม่สำเร็จ: ", err)
	}

	fmt.Println("เชื่อมต่อ Database สำเร็จแล้ว! 🎉")

	// Auto Migrate: สร้างตารางให้อัตโนมัติ ตาม Struct ที่เราเขียน
	DB.AutoMigrate(&model.News{}, &model.Activity{}, &model.Personnel{}, &model.User{})
	// ถ้ามี struct อื่นๆ ก็ใส่เพิ่มในวงเล็บได้เลยครับ
}

func SeedData() {
	var count int64
	DB.Model(&model.News{}).Count(&count)

}
