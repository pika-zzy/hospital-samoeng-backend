package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"hospitalbackend/database"
	models "hospitalbackend/models"

	"github.com/gin-gonic/gin"
)

// getOrCreatePopup — คืน record popup แถวแรก (singleton) ถ้ายังไม่มีให้สร้าง default ให้
func getOrCreatePopup() (models.Popup, error) {
	var popup models.Popup
	if err := database.DB.Order("id asc").First(&popup).Error; err == nil {
		return popup, nil
	}

	// ยังไม่มี record → สร้าง default (ปิดไว้ก่อน ไม่มีรูป)
	popup = models.Popup{Status: false, ImageURL: ""}
	if err := database.DB.Create(&popup).Error; err != nil {
		return models.Popup{}, err
	}
	return popup, nil
}

// GetPopup godoc
// @Summary  ดึงข้อมูล popup ของเว็บไซต์ (public)
// @Tags     popup
// @Produce  json
// @Success  200 {object} map[string]interface{}
// @Router   /popup [get]
func GetPopup(c *gin.Context) {
	popup, err := getOrCreatePopup()
	if err != nil {
		dbError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": popup})
}

// UpdatePopupStatus godoc
// @Summary  เปิด/ปิดการแสดง popup (เฉพาะ admin)
// @Tags     popup
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body object{status=bool} true "สถานะการแสดง"
// @Success  200 {object} map[string]interface{}
// @Router   /popup/status [patch]
func UpdatePopupStatus(c *gin.Context) {
	var body struct {
		Status *bool `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Status == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	popup, err := getOrCreatePopup()
	if err != nil {
		dbError(c, err)
		return
	}

	popup.Status = *body.Status
	if err := database.DB.Save(&popup).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": popup})
}

// UploadPopupImage godoc
// @Summary  อัปโหลด/เปลี่ยนรูป popup (เฉพาะ admin)
// @Tags     popup
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    image formData file true "รูปภาพ .jpg/.jpeg/.png"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Router   /popup/image [post]
func UploadPopupImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "กรุณาแนบรูปภาพ"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "อัปโหลดได้เฉพาะไฟล์รูปภาพ (.jpg/.jpeg/.png)"})
		return
	}

	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join("uploads/popup", newFileName)

	if err := os.MkdirAll("uploads/popup", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "สร้างโฟลเดอร์ไม่สำเร็จ"})
		return
	}

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
		return
	}

	popup, err := getOrCreatePopup()
	if err != nil {
		os.Remove(savePath)
		dbError(c, err)
		return
	}

	// เก็บ path รูปเก่าไว้ลบทีหลังเมื่อ save สำเร็จ (กันไฟล์ค้าง)
	oldImage := popup.ImageURL
	popup.ImageURL = "/uploads/popup/" + newFileName

	if err := database.DB.Save(&popup).Error; err != nil {
		os.Remove(savePath)
		dbError(c, err)
		return
	}

	// ลบรูปเก่าทิ้ง (best-effort — ไม่ให้ error บล็อก response)
	if oldImage != "" {
		os.Remove("." + oldImage)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": popup})
}
