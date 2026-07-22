package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"hospitalbackend/database"
	model "hospitalbackend/models"

	"github.com/gin-gonic/gin"
)

// GetHeroSlides godoc
// @Summary  list รูปสไลด์ hero หน้าแรก (เรียงตามลำดับ)
// @Tags     hero
// @Produce  json
// @Success  200 {object} map[string]interface{}
// @Router   /hero [get]
func GetHeroSlides(c *gin.Context) {
	var slides []model.HeroSlide
	// เรียงตามลำดับที่ admin กำหนด; ลำดับซ้ำกันให้ใช้ id เป็นตัวตัดสินให้ผลลัพธ์คงที่
	if err := database.DB.Order("sort_order asc, id asc").Find(&slides).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": slides})
}

// CreateHeroSlide godoc
// @Summary  เพิ่มรูปสไลด์ hero (admin)
// @Tags     hero
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    image formData file   true  "รูปภาพ .jpg/.jpeg/.png"
// @Param    order formData int    false "ลำดับการแสดง (ยิ่งน้อยยิ่งขึ้นก่อน)"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Router   /hero [post]
func CreateHeroSlide(c *gin.Context) {
	// CRIT-03: จำกัดขนาด body รวม (รูป + field) แล้ว parse ทันที
	if err := enforceUploadLimit(c, MaxImageBytes+maxFormOverhead); err != nil {
		uploadLimitError(c, err)
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "กรุณาแนบรูปภาพ"})
		return
	}

	ext, verr := validateUpload(file, allowedImageExt, MaxImageBytes)
	if verr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": verr.Error()})
		return
	}

	// ลำดับ: ไม่ส่งมา = ต่อท้ายรายการที่มีอยู่ (max + 1)
	order, err := strconv.Atoi(c.PostForm("order"))
	if err != nil {
		var maxOrder *int
		if err := database.DB.Model(&model.HeroSlide{}).
			Select("MAX(sort_order)").Scan(&maxOrder).Error; err != nil {
			dbError(c, err)
			return
		}
		if maxOrder == nil {
			order = 1
		} else {
			order = *maxOrder + 1
		}
	}

	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join("uploads/images/hero", newFileName)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
		return
	}
	imageURL := "/uploads/images/hero/" + newFileName

	slide := model.HeroSlide{ImageURL: imageURL, Order: order}
	if err := database.DB.Create(&slide).Error; err != nil {
		// DB พลาด → ลบไฟล์ที่เพิ่งเซฟทิ้ง ไม่ให้มีไฟล์ขยะค้าง
		os.Remove("." + imageURL)
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": slide})
}

// UpdateHeroSlideOrder godoc
// @Summary  แก้ลำดับรูปสไลด์ hero (admin)
// @Tags     hero
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id   path int true "hero slide id"
// @Param    body body map[string]int true "ลำดับใหม่ เช่น {\"order\": 2}"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /hero/{id} [put]
func UpdateHeroSlideOrder(c *gin.Context) {
	id := c.Param("id")

	var slide model.HeroSlide
	if err := database.DB.First(&slide, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบรูปสไลด์"})
		return
	}

	var body struct {
		Order *int `json:"order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Order == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	slide.Order = *body.Order
	if err := database.DB.Save(&slide).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": slide})
}

// DeleteHeroSlide godoc
// @Summary  ลบรูปสไลด์ hero (admin) — ห้ามลบจนเหลือ 0 รูป
// @Tags     hero
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "hero slide id"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{} "เหลือรูปเดียว ลบไม่ได้"
// @Failure  404 {object} map[string]interface{}
// @Router   /hero/{id} [delete]
func DeleteHeroSlide(c *gin.Context) {
	id := c.Param("id")

	var slide model.HeroSlide
	if err := database.DB.First(&slide, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบรูปสไลด์"})
		return
	}

	// กติกา: เมื่อมีรูปในระบบแล้ว ต้องเหลืออย่างน้อย 1 รูปเสมอ
	// (บังคับที่นี่ ไม่ใช่แค่ซ่อนปุ่มฝั่ง frontend — เรียก API ตรงก็ต้องโดนกัน)
	var count int64
	if err := database.DB.Model(&model.HeroSlide{}).Count(&count).Error; err != nil {
		dbError(c, err)
		return
	}
	if count <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ต้องมีรูปสไลด์อย่างน้อย 1 รูป ลบรูปสุดท้ายไม่ได้",
		})
		return
	}

	if err := database.DB.Delete(&model.HeroSlide{}, id).Error; err != nil {
		dbError(c, err)
		return
	}

	// ลบไฟล์จริงทิ้งด้วย (best-effort — ไม่ให้ error บล็อก response)
	if slide.ImageURL != "" {
		os.Remove("." + slide.ImageURL)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
