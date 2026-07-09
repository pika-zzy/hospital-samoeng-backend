package controllers

import (
	"errors"
	"fmt"
	"hospitalbackend/database"
	models "hospitalbackend/models"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ==================== MOIT ====================

// GetAllMoit godoc
// @Summary  list MOIT category ทั้งหมด
// @Tags     ita
// @Produce  json
// @Success  200 {object} map[string]interface{}
// @Router   /moit [get]
func GetAllMoit(c *gin.Context) {
	var moits []models.MoitCategory
	if err := database.DB.Order("name asc").Find(&moits).Error; err != nil {
		dbError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": moits})
}

// ==================== YEAR ====================

// GetYears godoc
// @Summary  list ปี ITA ทั้งหมด (เรียงใหม่→เก่า)
// @Tags     ita
// @Produce  json
// @Success  200 {object} map[string]interface{}
// @Router   /ita/years [get]
func GetYears(c *gin.Context) {
	var years []models.ITAYear
	// FIX: เดิมกลืน error เงียบ ๆ — DB พังก็ตอบ success:true พร้อม data ว่าง
	if err := database.DB.Order("year desc").Find(&years).Error; err != nil {
		dbError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": years})
}

// CreateYear godoc
// @Summary  สร้างปี ITA ใหม่ (เฉพาะ admin)
// @Tags     ita
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body object{year=int} true "ปี พ.ศ. 2500-2600"
// @Success  201 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{} "ปีไม่ถูกต้อง/ซ้ำ"
// @Router   /ita/years [post]
func CreateYear(c *gin.Context) {
	var body struct {
		Year int `json:"year"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	// validate range (ปี พ.ศ. — ปีงบประมาณไทย)
	if body.Year < 2500 || body.Year > 2600 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ปีไม่ถูกต้อง"})
		return
	}

	var exist models.ITAYear
	err := database.DB.Where("year = ?", body.Year).First(&exist).Error

	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ปีนี้มีอยู่แล้ว"})
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "เกิดข้อผิดพลาดในการค้นหา"})
		return
	}

	year := models.ITAYear{Year: body.Year}
	if err := database.DB.Create(&year).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": year})
}

// ==================== ITA ====================

// GetAllITA godoc
// @Summary  list ไฟล์ ITA (filter ตามปี + pagination)
// @Tags     ita
// @Produce  json
// @Param    year_id query int false "filter ตาม year id"
// @Param    page    query int false "หน้า (default 1)"
// @Param    limit   query int false "ต่อหน้า (default 20, max 100)"
// @Success  200 {object} map[string]interface{} "data + meta{total,page,limit}"
// @Router   /ita [get]
func GetAllITA(c *gin.Context) {
	yearIDStr := c.Query("year_id")

	// pagination
	page,
		_ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// count query แยกต่างหาก ให้ filter ตรงกัน
	countQuery := database.DB.Model(&models.ITA{})
	if yearIDStr != "" {
		countQuery = countQuery.Where("year_id = ?", yearIDStr)
	}
	var total int64
	countQuery.Count(&total)

	// data query
	dataQuery := database.DB.
		Preload("Item.Topic.Moit").
		Preload("Year")

	if yearIDStr != "" {
		dataQuery = dataQuery.Where("year_id = ?", yearIDStr)
	}

	var itas []models.ITA
	if err := dataQuery.Limit(limit).Offset(offset).Find(&itas).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    itas,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// DeleteITA godoc
// @Summary  ลบไฟล์ ITA (เฉพาะ admin)
// @Tags     ita
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "ITA file id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{} "ไม่พบไฟล์"
// @Router   /ita/{id} [delete]
func DeleteITA(c *gin.Context) {
	id := c.Param("id")

	var ita models.ITA
	if err := database.DB.First(&ita, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบไฟล์"})
		return
	}

	if err := database.DB.Delete(&models.ITA{}, id).Error; err != nil {
		dbError(c, err)
		return
	}

	// ลบไฟล์จริงทิ้งด้วย (best-effort — ไม่ให้ error บล็อก response)
	if ita.FileURL != "" {
		os.Remove("." + ita.FileURL)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UploadITA godoc
// @Summary  อัปโหลดไฟล์ PDF เข้า MOIT item (เฉพาะ admin)
// @Tags     ita
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    item_id formData int    true  "MOIT item id"
// @Param    year_id formData int    true  "ITA year id (ต้องตรงกับ item)"
// @Param    title   formData string false "ชื่อเอกสาร (ไม่ใส่ = ใช้ชื่อไฟล์)"
// @Param    file    formData file   true  "ไฟล์ PDF"
// @Success  201 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Router   /ita/upload [post]
func UploadITA(c *gin.Context) {
	itemIDStr := c.PostForm("item_id")
	title := c.PostForm("title")
	yearIDStr := c.PostForm("year_id")

	// parse IDs
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil || itemID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "item_id ไม่ถูกต้อง"})
		return
	}

	yearID, err := strconv.ParseUint(yearIDStr, 10, 64)
	if err != nil || yearID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "year_id ไม่ถูกต้อง"})
		return
	}

	// validate item + relation ในครั้งเดียว
	var item models.MoitItem
	if err := database.DB.Preload("Topic.Moit").First(&item, itemID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ไม่พบ item"})
		return
	}

	// ตรวจ year ตรงกับ item หรือเปล่า
	if item.Topic.Moit.YearID != uint(yearID) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "year ไม่ตรงกับ item"})
		return
	}

	// รับไฟล์
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "กรุณาแนบไฟล์"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "อัปโหลดได้เฉพาะ PDF"})
		return
	}

	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join("uploads/ita", newFileName)

	// สร้างโฟลเดอร์อัตโนมัติ
	if err := os.MkdirAll("uploads/ita", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "สร้างโฟลเดอร์ไม่สำเร็จ"})
		return
	}

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
		return
	}

	if title == "" {
		title = file.Filename
	}

	ita := models.ITA{
		ItemID:  uint(itemID),
		Title:   title,
		FileURL: "/uploads/ita/" + newFileName,
		YearID:  uint(yearID),
	}

	if err := database.DB.Create(&ita).Error; err != nil {
		// DB fail → ลบไฟล์ทิ้ง
		os.Remove(savePath)
		dbError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": ita})
}

// UpdateITA godoc
// @Summary  แก้ไขเอกสาร ITA — เปลี่ยนชื่อ และ/หรือ เปลี่ยนไฟล์ PDF (เฉพาะ admin)
// @Tags     ita
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    id    path     int    true  "ITA id"
// @Param    title formData string false "ชื่อเอกสารใหม่"
// @Param    file  formData file   false "ไฟล์ PDF ใหม่ (ไม่แนบ = ใช้ไฟล์เดิม)"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /ita/{id} [put]
func UpdateITA(c *gin.Context) {
	id := c.Param("id")

	var ita models.ITA
	if err := database.DB.First(&ita, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบไฟล์"})
		return
	}

	// ชื่อเอกสาร (ถ้าส่งมา)
	if title := c.PostForm("title"); title != "" {
		ita.Title = title
	}

	// ไฟล์ใหม่ (optional) — แนบมา = แทนที่ไฟล์เดิม
	file, err := c.FormFile("file")
	if err == nil {
		ext := filepath.Ext(file.Filename)
		if ext != ".pdf" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "อัปโหลดได้เฉพาะ PDF"})
			return
		}

		newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join("uploads/ita", newFileName)

		if err := os.MkdirAll("uploads/ita", os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "สร้างโฟลเดอร์ไม่สำเร็จ"})
			return
		}

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
			return
		}

		oldFileURL := ita.FileURL
		ita.FileURL = "/uploads/ita/" + newFileName

		// บันทึก DB ก่อน — ถ้าพัง ลบไฟล์ใหม่ทิ้ง คงไฟล์เดิมไว้
		if err := database.DB.Save(&ita).Error; err != nil {
			os.Remove(savePath)
			dbError(c, err)
			return
		}

		// สำเร็จ → ลบไฟล์เก่าทิ้ง (best-effort)
		if oldFileURL != "" {
			os.Remove("." + oldFileURL)
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": ita})
		return
	}

	// ไม่มีไฟล์ใหม่ — บันทึกเฉพาะ field ที่แก้
	if err := database.DB.Save(&ita).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": ita})
}
