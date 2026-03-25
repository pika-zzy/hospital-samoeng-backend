package main

import (
	"hospitalbackend/database"
	"hospitalbackend/routes" // ⚠️ เปลี่ยนตามชื่อ module ถ้าไม่ตรง
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// debug ดูว่าอ่านได้ไหม
	println("JWT_SECRET:", os.Getenv("JWT_SECRET"))
	database.ConnectDB()
	r := routes.SetupRouter()
	r.Run("0.0.0.0:8080")
}
