package controllers

import (
	"fmt"
	"hospitalbackend/database"
	model "hospitalbackend/models"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateITA(c *gin.Context) {
	//รับค่าข้อความ (Text Fields) จาก FromData
	title := c.PostForm("title")
	year := c.PostForm("year")
	var fileURL string = " " // กำหนดค่าเริ่มต้นเป็นว่าง
	// รับไฟล์ (File)
	file, err := c.FormFile("file")
	if err == nil {
		// ถ้ามีไฟล์ถูกอัปโหลดมา ให้บันทึกไฟล์ลงในโฟลเดอร์ "uploads"
		ext := filepath.Ext(file.Filename) // ดึงนามสกุลไฟล์
		//สั่ง บันทึกไฟล์ลงในโฟลเดอร์ "uploads"
		if ext != ".pdf" {
			c.JSON(400, gin.H{"success": false, "message": "อัพโหลดได้เฉพาะไฟล์ PDF เท่านั้น"})
			return
		}
		newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext) // สร้างชื่อไฟล์ใหม่แบบไม่ซ้ำ
		savePath := filepath.Join("uploads/file/ita", newFileName)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
			return
		}

		// กำหนด URL ของไฟล์ที่ถูกอัปโหลด (สมมติว่าไฟล์จะถูกเสิร์ฟจาก /uploads/)
		fileURL = "/uploads/file/ita/" + newFileName
	}

	// สร้างไฟล์
	itanew := model.ITA{
		Title:   title,
		Year:    year,
		FileURL: fileURL, // กำหนด URL ของไฟล์ที่ถูกอัปโหลด
	}
	result := database.DB.Create(&itanew)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": result.Error.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    itanew,
	})

}

func GetAllITA(c *gin.Context) {
	var ita []model.ITA
	result := database.DB.Find(&ita)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ita,
	})
}

func GetITAByID(c *gin.Context) {
	id := c.Param("id")
	var ita model.ITA
	result := database.DB.Where("id = ?", id).First(&ita)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "ita not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ita,
	})
}
