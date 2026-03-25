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

func GetAllActivities(c *gin.Context) {
	var activities []model.Activity
	result := database.DB.Find(&activities)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
	})
}

func GetActivityByID(c *gin.Context) {

	id := c.Param("id")

	var activity model.Activity
	result := database.DB.Where("id = ?", id).First(&activity)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "activity not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activity,
	})
}

func CreateActivity(c *gin.Context) {

	title := c.PostForm("title")
	description := c.PostForm("description")
	startDate := c.PostForm("startDate")
	endDate := c.PostForm("endDate")

	var imgURL string = ""

	file, err := c.FormFile("image")

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

		savePath := filepath.Join("uploads/images/activity", newFileName)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "บันทึกไฟล์ไม่สำเร็จ",
			})
			return
		}

		imgURL = "/uploads/images/activity/" + newFileName
	}

	activity := model.Activity{
		Title:       title,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		ImgURL:      imgURL,
	}

	result := database.DB.Create(&activity)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activity,
	})
}
