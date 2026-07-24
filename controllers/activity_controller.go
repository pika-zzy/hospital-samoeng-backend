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

// GetAllActivities godoc
// @Summary  list กิจกรรมทั้งหมด
// @Tags     activities
// @Produce  json
// @Success  200 {object} map[string]interface{}
// @Router   /activities [get]
func GetAllActivities(c *gin.Context) {
	var activities []model.Activity
	result := database.DB.Find(&activities)

	if result.Error != nil {
		dbError(c, result.Error)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
	})
}

// GetActivityByID godoc
// @Summary  ดูกิจกรรมรายตัว
// @Tags     activities
// @Produce  json
// @Param    id path int true "activity id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /activities/{id} [get]
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

// CreateActivity godoc
// @Summary  สร้างกิจกรรมใหม่ (แนบรูปได้)
// @Tags     activities
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    title       formData string true  "ชื่อกิจกรรม"
// @Param    description formData string false "รายละเอียด"
// @Param    startDate   formData string false "วันเริ่ม"
// @Param    endDate     formData string false "วันจบ"
// @Param    image       formData file   false "รูปภาพ .jpg/.jpeg/.png"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{} "นามสกุลไฟล์ไม่ถูกต้อง"
// @Router   /activities [post]
func CreateActivity(c *gin.Context) {

	// CRIT-03: จำกัดขนาด body รวม (รูป + field) แล้ว parse ทันที
	if err := enforceUploadLimit(c, MaxImageBytes+maxFormOverhead); err != nil {
		uploadLimitError(c, err)
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	startDate := c.PostForm("startDate")
	endDate := c.PostForm("endDate")

	var imgURL string = ""

	file, err := c.FormFile("image")

	if err == nil {

		ext, verr := validateUpload(file, allowedImageExt, MaxImageBytes)
		if verr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": verr.Error()})
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
		dbError(c, result.Error)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activity,
	})
}

// UpdateActivity godoc
// @Summary  แก้ไขกิจกรรม (เปลี่ยนรูปได้, ไม่แนบรูป = ใช้รูปเดิม)
// @Tags     activities
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    id          path     int    true  "activity id"
// @Param    title       formData string true  "ชื่อกิจกรรม"
// @Param    description formData string false "รายละเอียด"
// @Param    startDate   formData string false "วันเริ่ม"
// @Param    endDate     formData string false "วันจบ"
// @Param    image       formData file   false "รูปภาพใหม่ .jpg/.jpeg/.png"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /activities/{id} [put]
func UpdateActivity(c *gin.Context) {
	id := c.Param("id")

	// CRIT-03: จำกัดขนาด body รวม (รูป + field) แล้ว parse ทันที
	if err := enforceUploadLimit(c, MaxImageBytes+maxFormOverhead); err != nil {
		uploadLimitError(c, err)
		return
	}

	var activity model.Activity
	if err := database.DB.First(&activity, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบกิจกรรม"})
		return
	}

	// ถ้ามีรูปใหม่แนบมา → บันทึกรูปใหม่ แล้วเก็บ path รูปเก่าไว้ลบทีหลัง
	oldImage := activity.ImgURL
	newImage := ""
	if file, err := c.FormFile("image"); err == nil {
		ext, verr := validateUpload(file, allowedImageExt, MaxImageBytes)
		if verr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": verr.Error()})
			return
		}
		newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join("uploads/images/activity", newFileName)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
			return
		}
		newImage = "/uploads/images/activity/" + newFileName
	}

	activity.Title = c.PostForm("title")
	activity.Description = c.PostForm("description")
	activity.StartDate = c.PostForm("startDate")
	activity.EndDate = c.PostForm("endDate")
	if newImage != "" {
		activity.ImgURL = newImage
	}

	if err := database.DB.Save(&activity).Error; err != nil {
		// DB พลาด → ลบรูปใหม่ที่เพิ่งบันทึกทิ้ง คงรูปเดิมไว้
		if newImage != "" {
			os.Remove("." + newImage)
		}
		dbError(c, err)
		return
	}

	// เปลี่ยนรูปสำเร็จ → ลบรูปเก่าทิ้ง (best-effort)
	if newImage != "" && oldImage != "" {
		os.Remove("." + oldImage)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": activity})
}

// DeleteActivity godoc
// @Summary  ลบกิจกรรม (เฉพาะ admin/employee)
// @Tags     activities
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "activity id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /activities/{id} [delete]
func DeleteActivity(c *gin.Context) {
	id := c.Param("id")

	var activity model.Activity
	if err := database.DB.First(&activity, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบกิจกรรม"})
		return
	}

	if err := database.DB.Delete(&model.Activity{}, id).Error; err != nil {
		dbError(c, err)
		return
	}

	// ลบไฟล์รูปทิ้งด้วย (best-effort — ไม่ให้ error บล็อก response)
	if activity.ImgURL != "" {
		os.Remove("." + activity.ImgURL)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
