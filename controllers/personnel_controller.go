package controllers

import (
	"fmt"
	"hospitalbackend/database"
	model "hospitalbackend/models"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetAllPersonnel godoc
// @Summary  list บุคลากรทั้งหมด
// @Tags     personnel
// @Produce  json
// @Success  200 {object} map[string]interface{}
// @Router   /personnel [get]
func GetAllPersonnel(c *gin.Context) {
	var personnel []model.Personnel
	result := database.DB.Find(&personnel)

	if result.Error != nil {
		dbError(c, result.Error)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    personnel,
	})
}

// GetPersonnelByID godoc
// @Summary  ดูบุคลากรรายคน
// @Tags     personnel
// @Produce  json
// @Param    id path int true "personnel id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /personnel/{id} [get]
func GetPersonnelByID(c *gin.Context) {
	id := c.Param("id")
	var personnel model.Personnel
	result := database.DB.Where("id = ?", id).First(&personnel)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "personnel not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    personnel,
	})

}

// AddNewPersonnal godoc
// @Summary  เพิ่มบุคลากรใหม่ (แนบรูปได้)
// @Tags     personnel
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    prefix   formData string true  "คำนำหน้า"
// @Param    name     formData string true  "ชื่อ"
// @Param    lastname formData string true  "นามสกุล"
// @Param    uid      formData int    true  "เลขประจำตัว"
// @Param    role     formData string true  "ฝ่ายงาน/แผนก"
// @Param    position formData string false "ตำแหน่ง"
// @Param    image    formData file   false "รูปภาพ .jpg/.jpeg/.png"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Router   /personnel [post]
func AddNewPersonnal(c *gin.Context) {
	prefix := c.PostForm("prefix")
	name := c.PostForm("name")
	lastname := c.PostForm("lastname")
	uidstr := c.PostForm("uid")
	role := c.PostForm("role")
	position := c.PostForm("position")
	var imgURL string = ""

	//แปลง uid ให้เป็น int
	uid, err := strconv.Atoi(uidstr)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "message": "uid must be number"})
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

		// FIX: เติม / นำหน้าให้ตรง convention เดียวกับ controller อื่น (เดิมไม่มี ทำให้ URL ฝั่ง frontend เพี้ยน)
		imgURL = "/uploads/image/personnel/" + imgNewFilename
	}

	personnel := model.Personnel{
		Prefix:   prefix,
		Name:     name,
		Lastname: lastname,
		Uid:      uid,
		Role:     role,
		Position: position,
		ImgURL:   imgURL,
	}

	result := database.DB.Create(&personnel)

	if result.Error != nil {
		// FIX: เดิมไม่มี return ทำให้เขียน response ซ้อน 2 ก้อน (500 ตามด้วย success:true)
		dbError(c, result.Error)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    personnel,
	})
}

// UpdatePersonnel godoc
// @Summary  แก้ไขบุคลากร (เปลี่ยนรูปได้, ไม่แนบรูป = ใช้รูปเดิม)
// @Tags     personnel
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    id       path     int    true  "personnel id"
// @Param    prefix   formData string true  "คำนำหน้า"
// @Param    name     formData string true  "ชื่อ"
// @Param    lastname formData string true  "นามสกุล"
// @Param    uid      formData int    true  "เลขประจำตัว"
// @Param    role     formData string true  "ฝ่ายงาน/แผนก"
// @Param    position formData string false "ตำแหน่ง"
// @Param    image    formData file   false "รูปภาพใหม่ .jpg/.jpeg/.png"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /personnel/{id} [put]
func UpdatePersonnel(c *gin.Context) {
	id := c.Param("id")

	var personnel model.Personnel
	if err := database.DB.First(&personnel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบบุคลากร"})
		return
	}

	uid, err := strconv.Atoi(c.PostForm("uid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "uid must be number"})
		return
	}

	// ถ้ามีรูปใหม่แนบมา → บันทึกรูปใหม่ แล้วเก็บ path รูปเก่าไว้ลบทีหลัง
	oldImage := personnel.ImgURL
	newImage := ""
	if file, err := c.FormFile("image"); err == nil {
		ext := filepath.Ext(file.Filename)
		if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "อัพโหลดได้เฉพาะไฟล์รูปภาพ"})
			return
		}

		imgNewFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join("uploads/image/personnel", imgNewFilename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "อัพโหลดรูปภาพไม่สำเร็จ"})
			return
		}
		newImage = "/uploads/image/personnel/" + imgNewFilename
	}

	personnel.Prefix = c.PostForm("prefix")
	personnel.Name = c.PostForm("name")
	personnel.Lastname = c.PostForm("lastname")
	personnel.Uid = uid
	personnel.Role = c.PostForm("role")
	personnel.Position = c.PostForm("position")
	if newImage != "" {
		personnel.ImgURL = newImage
	}

	if err := database.DB.Save(&personnel).Error; err != nil {
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

	c.JSON(http.StatusOK, gin.H{"success": true, "data": personnel})
}

// DeletePersonnel godoc
// @Summary  ลบบุคลากร (เฉพาะ admin/employee)
// @Tags     personnel
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "personnel id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /personnel/{id} [delete]
func DeletePersonnel(c *gin.Context) {
	id := c.Param("id")

	var personnel model.Personnel
	if err := database.DB.First(&personnel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบบุคลากร"})
		return
	}

	if err := database.DB.Delete(&model.Personnel{}, id).Error; err != nil {
		dbError(c, err)
		return
	}

	// ลบไฟล์รูปทิ้งด้วย (best-effort — ไม่ให้ error บล็อก response)
	if personnel.ImgURL != "" {
		os.Remove("." + personnel.ImgURL)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
