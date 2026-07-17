package controllers

import (
	model "hospitalbackend/models"
	"net/http"

	"hospitalbackend/database"

	"fmt"
	"os"
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

// UpdateNews godoc
// @Summary  แก้ไขข่าว (เปลี่ยนไฟล์/รูปได้, ไม่แนบ = ใช้ของเดิม)
// @Tags     news
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    id          path     int    true  "news id"
// @Param    title       formData string true  "หัวข้อข่าว"
// @Param    description formData string false "รายละเอียด"
// @Param    date        formData string false "วันที่"
// @Param    type        formData string false "ประเภทข่าว"
// @Param    file        formData file   false "ไฟล์ PDF ใหม่"
// @Param    image       formData file   false "รูปภาพใหม่ .jpg/.jpeg/.png"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /news/{id} [put]
func UpdateNews(c *gin.Context) {
	id := c.Param("id")

	var news model.News
	if err := database.DB.First(&news, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบข่าว"})
		return
	}

	// ไฟล์ใหม่ (ถ้าแนบมา) — เก็บ path เก่าไว้ลบหลังบันทึกสำเร็จ
	oldFile := news.FileURL
	oldImage := news.ImgURL
	newFile := ""
	newImage := ""

	if file, err := c.FormFile("file"); err == nil {
		ext := filepath.Ext(file.Filename)
		if ext != ".pdf" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "อัพโหลดได้เฉพาะไฟล์ PDF เท่านั้น"})
			return
		}
		newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join("uploads/file/news", newFileName)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
			return
		}
		newFile = "/uploads/file/news/" + newFileName
	}

	if img, err := c.FormFile("image"); err == nil {
		ext := filepath.Ext(img.Filename)
		if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "อัพโหลดได้เฉพาะไฟล์รูปภาพ"})
			return
		}
		newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join("uploads/images/news", newFileName)
		if err := c.SaveUploadedFile(img, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
			return
		}
		newImage = "/uploads/images/news/" + newFileName
	}

	news.Title = c.PostForm("title")
	news.Description = c.PostForm("description")
	news.Date = c.PostForm("date")
	news.Type = c.PostForm("type")
	if newFile != "" {
		news.FileURL = newFile
	}
	if newImage != "" {
		news.ImgURL = newImage
	}

	if err := database.DB.Save(&news).Error; err != nil {
		// DB พลาด → ลบไฟล์ใหม่ที่เพิ่งบันทึกทิ้ง (ไม่ให้มีไฟล์ขยะค้าง)
		if newFile != "" {
			os.Remove("." + newFile)
		}
		if newImage != "" {
			os.Remove("." + newImage)
		}
		dbError(c, err)
		return
	}

	// บันทึกสำเร็จ → ลบไฟล์เก่าที่ถูกแทนที่ทิ้ง (best-effort)
	if newFile != "" && oldFile != "" {
		os.Remove("." + oldFile)
	}
	if newImage != "" && oldImage != "" {
		os.Remove("." + oldImage)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": news})
}

// DeleteNews godoc
// @Summary  ลบข่าว (เฉพาะ admin/employee)
// @Tags     news
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "news id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /news/{id} [delete]
func DeleteNews(c *gin.Context) {
	id := c.Param("id")

	var news model.News
	if err := database.DB.First(&news, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบข่าว"})
		return
	}

	if err := database.DB.Delete(&model.News{}, id).Error; err != nil {
		dbError(c, err)
		return
	}

	// ลบไฟล์แนบทิ้งด้วย (best-effort — ไม่ให้ error บล็อก response)
	if news.FileURL != "" {
		os.Remove("." + news.FileURL)
	}
	if news.ImgURL != "" {
		os.Remove("." + news.ImgURL)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
