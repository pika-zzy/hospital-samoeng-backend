package controllers

import (
	"fmt"
	"hospitalbackend/database"
	model "hospitalbackend/models"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateITA(c *gin.Context) {
	title := c.PostForm("title")
	year := c.PostForm("year")

	//  validate
	if title == "" || year == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "title and year are required",
		})
		return
	}

	//  กัน year แปลก
	if len(year) != 4 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid year",
		})
		return
	}

	//  รับไฟล์
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "file is required",
		})
		return
	}

	//  จำกัดเฉพาะ PDF
	ext := filepath.Ext(file.Filename)
	if ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "only pdf allowed",
		})
		return
	}

	//  ตั้งชื่อไฟล์ใหม่ (กันชน)
	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	//  path แยกปี
	yearPath := filepath.Join("uploads/file/ita", year)

	//  สร้างโฟลเดอร์อัตโนมัติ
	if err := os.MkdirAll(yearPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "cannot create folder",
		})
		return
	}

	//  path จริง
	savePath := filepath.Join(yearPath, newFileName)

	//  save ไฟล์
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "upload failed",
		})
		return
	}

	// URL
	fileURL := "/uploads/file/ita/" + year + "/" + newFileName

	//  save DB
	ita := model.ITA{
		Title:   title,
		Year:    year,
		FileURL: fileURL,
	}

	if err := database.DB.Create(&ita).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "save db failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ita,
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
