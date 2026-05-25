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

// GET /moit
func GetAllMoit(c *gin.Context) {
	var moits []models.MoitCategory
	if err := database.DB.Order("name asc").Find(&moits).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": moits})
}

// ==================== YEAR ====================

// GET /ita/years
func GetYears(c *gin.Context) {
	var years []models.ITAYear
	database.DB.Order("year desc").Find(&years)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": years})
}

// POST /ita/years
func CreateYear(c *gin.Context) {
	var body struct {
		Year int `json:"year"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	// validate range
	if body.Year < 2000 || body.Year > 2100 {
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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": year})
}

// ==================== ITA ====================

// GET /ita
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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
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

// POST /ita/upload
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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": ita})
}
