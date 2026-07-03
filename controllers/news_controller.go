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

// GetAllNews godoc
// @Summary  list ข่าวทั้งหมด
// @Tags     news
// @Produce  json
// @Success  200 {object} map[string]interface{}
// @Router   /news [get]
func GetAllNews(c *gin.Context) {
	var news []model.News
	result := database.DB.Find(&news)

	if result.Error != nil {
		dbError(c, result.Error)
		return
	}
	// FIX: ตัด message "ยังไม่มีข้อมูลในขณะนี้" ที่ติดมากับ response สำเร็จทุกครั้ง (frontend ไม่ได้ใช้)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    news,
	})

}

// GetNewsByID godoc
// @Summary  ดูข่าวรายตัว
// @Tags     news
// @Produce  json
// @Param    id path int true "news id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /news/{id} [get]
func GetNewsByID(c *gin.Context) {
	id := c.Param("id")
	// หาข่าวที่ ID ตรงกัน (แบบบ้านๆ ไปก่อน)
	var news model.News
	result := database.DB.Where("id = ?", id).First(&news)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "news not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    news,
	})
}

// CreateNews godoc
// @Summary  สร้างข่าวใหม่ (แนบไฟล์ PDF + รูปได้)
// @Tags     news
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    title       formData string true  "หัวข้อข่าว"
// @Param    description formData string false "รายละเอียด"
// @Param    date        formData string false "วันที่"
// @Param    type        formData string false "ประเภทข่าว"
// @Param    file        formData file   false "ไฟล์ PDF"
// @Param    image       formData file   false "รูปภาพ .jpg/.jpeg/.png"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{} "นามสกุลไฟล์ไม่ถูกต้อง"
// @Router   /news [post]
func CreateNews(c *gin.Context) {
	//รับค่าข้อความ (Text Fields) จาก FromData
	title := c.PostForm("title")
	description := c.PostForm("description")
	date := c.PostForm("date")
	typeNews := c.PostForm("type")
	var imgURL string = ""
	// FIX: เดิมเป็น " " (มีช่องว่าง) ไม่ใช่สตริงว่าง — frontend เช็ค truthy เลยโชว์ปุ่มดาวน์โหลดทั้งที่ไม่มีไฟล์
	var fileURL string = ""
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
		savePath := filepath.Join("uploads/file/news", newFileName)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
			return
		}

		// กำหนด URL ของไฟล์ที่ถูกอัปโหลด (สมมติว่าไฟล์จะถูกเสิร์ฟจาก /uploads/)
		fileURL = "/uploads/file/news/" + newFileName
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
		dbError(c, result.Error)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    news,
	})
}
