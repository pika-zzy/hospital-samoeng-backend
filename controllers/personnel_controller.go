package controllers

import (
	"fmt"
	"hospitalbackend/database"
	model "hospitalbackend/models"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetAllPersonnel(c *gin.Context) {
	var personnel []model.Personnel
	result := database.DB.Find(&personnel)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": result.Error.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    personnel,
	})
}

func GetPersonnelByID(c *gin.Context) {
	id := c.Param("id")
	var personnel model.Personnel
	result := database.DB.Where("id = ?", id).First(&personnel)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "personnel not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    personnel,
	})

}

func AddNewPersonnal(c *gin.Context) {
	prefix := c.PostForm("prefix")
	name := c.PostForm("name")
	lastname := c.PostForm("lastname")
	uidstr := c.PostForm("uid")
	role := c.PostForm("role")
	var imgURL string = ""

	//แปลง uid ให้เป็น int
	uid, err := strconv.Atoi(uidstr)
	if err != nil {
		c.JSON(400, gin.H{"message": "uid must be number"})
		return
	}

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

		imgNewFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join("uploads/image/personnel", imgNewFilename)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "อัพโหลดรูปภาพไม่สำเร็จ",
			})
			return
		}

		imgURL = "uploads/image/personnel/" + imgNewFilename
	}

	personnel := model.Personnel{
		Prefix:   prefix,
		Name:     name,
		Lastname: lastname,
		Uid:      uid,
		Role:     role,
		ImgURL:   imgURL,
	}

	result := database.DB.Create(&personnel)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": result.Error.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    personnel,
	})
}
