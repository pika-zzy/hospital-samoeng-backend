package controllers

import (
	model "hospitalbackend/models"
	"net/http"

	"hospitalbackend/database"

	"fmt"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func GetAllNews(c *gin.Context) {
	var news []model.News
	result := database.DB.Find(&news)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": result.Error.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    news,
		"message": "ยังไม่มีข้อมูลในขณะนี้",
	})

}

func GetNewsByID(c *gin.Context) {
	id := c.Param("id")
	// หาข่าวที่ ID ตรงกัน (แบบบ้านๆ ไปก่อน)
	var news model.News
	result := database.DB.Where("id = ?", id).First(&news)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "news not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    news,
	})
}

func CreateNews(c *gin.Context) {
	//รับค่าข้อความ (Text Fields) จาก FromData
	title := c.PostForm("title")
	description := c.PostForm("description")
	date := c.PostForm("date")
	typeNews := c.PostForm("type")
	var imgURL string = ""
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
		savePath := filepath.Join("uploads/file", newFileName)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
			return
		}

		// กำหนด URL ของไฟล์ที่ถูกอัปโหลด (สมมติว่าไฟล์จะถูกเสิร์ฟจาก /uploads/)
		fileURL = "/uploads/file/" + newFileName
	}

	file, err = c.FormFile("image")

	if err == nil {

		ext := filepath.Ext(file.Filename)

		if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
			c.JSON(400, gin.H{
				"success": false,
				"message": "อัพโหลดได้เฉพาะไฟล์รูปภาพ",
			})
			return
		}

		newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

		savePath := filepath.Join("uploads/images/news", newFileName)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "บันทึกไฟล์ไม่สำเร็จ",
			})
			return
		}

		imgURL = "/uploads/images/news/" + newFileName
	}

	// สร้างข่าวใหม่
	news := model.News{
		Title:       title,
		Description: description,
		Date:        date,
		Type:        typeNews,
		FileURL:     fileURL, // กำหนด URL ของไฟล์ที่ถูกอัปโหลด
		ImgURL:      imgURL,
	}
	result := database.DB.Create(&news)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": result.Error.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    news,
	})
}
