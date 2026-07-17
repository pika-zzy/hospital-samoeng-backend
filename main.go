package main

import (
	"hospitalbackend/database"
	"hospitalbackend/routes"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// @title           Hospital Samoeng API
// @version         1.0
// @description     REST API หลังบ้านเว็บไซต์โรงพยาบาลสะเมิง (ข่าว กิจกรรม บุคลากร ITA/MOIT ผู้ใช้งาน สิทธิ์เมนู)
// @host            localhost:8080
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     ใส่ค่าเป็น "Bearer <token>" (ได้ token จาก POST /login)
func main() {
	// โหลด .env แบบ optional: ตอน dev มีไฟล์ก็อ่านค่าเข้ามา
	// ตอน deploy (เช่น Dokploy) ไม่มีไฟล์ .env แต่ env ถูกฉีดมาที่ระดับ OS อยู่แล้ว
	// จึงไม่ควร fatal เพราะหาไฟล์ไม่เจอ — ตัวที่ต้องมีจริงถูกเช็คด้านล่างแยกต่างหาก
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file loaded, using environment variables from the system")
	}

	// FIX: เช็คว่ามี JWT_SECRET โดยไม่พิมพ์ค่า secret ลง log (ของเดิม println ค่าออกมาตรง ๆ)
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET is not set")
	}
	database.ConnectDB()
	r := routes.SetupRouter()
	r.Run("0.0.0.0:8080")
}
