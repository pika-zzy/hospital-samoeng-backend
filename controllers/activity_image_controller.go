package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"hospitalbackend/database"
	model "hospitalbackend/models"

	"github.com/gin-gonic/gin"
)

// รูปอัลบั้มกิจกรรมเก็บโฟลเดอร์เดียวกับรูปปก — เสิร์ฟที่ /uploads/images/activity/...
const activityImageDir = "uploads/images/activity"

// countActivityImages นับ "จำนวนรูปที่ผู้ใช้เห็นจริง" ของกิจกรรมหนึ่ง
// = รูปปก (ถ้ามี) + รูปในอัลบั้ม — ใช้เทียบกับเพดาน model.MaxActivityImages
func countActivityImages(a *model.Activity) (int64, error) {
	var n int64
	if err := database.DB.Model(&model.ActivityImage{}).
		Where("activity_id = ?", a.ID).Count(&n).Error; err != nil {
		return 0, err
	}
	if a.ImgURL != "" {
		n++
	}
	return n, nil
}

// AddActivityImages godoc
// @Summary  เพิ่มรูปเข้าอัลบั้มกิจกรรม (ครั้งละหลายรูปได้)
// @Tags     activities
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    id     path     int   true "activity id"
// @Param    images formData file  true "รูป .jpg/.jpeg/.png (ส่งซ้ำ field เดิมได้หลายไฟล์)"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{} "ไฟล์ไม่ถูกต้อง หรือรูปเกินเพดาน"
// @Failure  404 {object} map[string]interface{}
// @Router   /activities/{id}/images [post]
//
// เพดาน: รูปปก + อัลบั้ม รวมกันต้องไม่เกิน model.MaxActivityImages (8)
// ถ้าส่งมาเกินโควตาที่เหลือ → ตอบ 400 พร้อมบอกว่าเหลือกี่รูป **ไม่บันทึกสักไฟล์**
// (ตัดสินใจแบบ all-or-nothing เพื่อไม่ให้ admin เดาไม่ออกว่ารูปไหนเข้าไม่เข้า)
func AddActivityImages(c *gin.Context) {
	id := c.Param("id")

	// เพดาน body = จำนวนรูปที่เพิ่มได้มากสุดในคำขอเดียว + เผื่อ field อื่น
	maxBody := MaxImageBytes*int64(model.MaxActivityImages-1) + maxFormOverhead
	if err := enforceUploadLimit(c, maxBody); err != nil {
		uploadLimitError(c, err)
		return
	}

	var activity model.Activity
	if err := database.DB.First(&activity, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบกิจกรรม"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}
	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ไม่ได้แนบรูปมา"})
		return
	}

	used, err := countActivityImages(&activity)
	if err != nil {
		dbError(c, err)
		return
	}
	left := int64(model.MaxActivityImages) - used
	if left <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("กิจกรรมนี้มีรูปครบ %d รูปแล้ว ลบรูปเดิมก่อนถึงจะเพิ่มได้", model.MaxActivityImages),
		})
		return
	}
	if int64(len(files)) > left {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("เพิ่มได้อีกไม่เกิน %d รูป (เพดาน %d รูปต่อกิจกรรม นับรวมรูปปกแล้ว) แต่แนบมา %d รูป",
				left, model.MaxActivityImages, len(files)),
		})
		return
	}

	// ตรวจไฟล์ให้ครบทุกใบก่อน แล้วค่อยเซฟ — ไฟล์เสียใบเดียวไม่ทิ้งขยะไว้บนดิสก์
	exts := make([]string, len(files))
	for i, f := range files {
		ext, verr := validateUpload(f, allowedImageExt, MaxImageBytes)
		if verr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": fmt.Sprintf("ไฟล์ \"%s\": %s", filepath.Base(f.Filename), verr.Error()),
			})
			return
		}
		exts[i] = ext
	}

	// ต่อท้ายอัลบั้มเดิม — เอา sort_order สูงสุดที่มีอยู่มาเป็นจุดเริ่ม
	var maxOrder int
	database.DB.Model(&model.ActivityImage{}).
		Where("activity_id = ?", activity.ID).
		Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)

	saved := make([]string, 0, len(files))
	rows := make([]model.ActivityImage, 0, len(files))
	for i, f := range files {
		newFileName := fmt.Sprintf("%d-%d%s", time.Now().UnixNano(), i, exts[i])
		savePath := filepath.Join(activityImageDir, newFileName)
		if err := c.SaveUploadedFile(f, savePath); err != nil {
			for _, p := range saved { // เซฟไม่ครบ → เก็บกวาดไฟล์ที่เพิ่งเขียนไป
				os.Remove("." + p)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
			return
		}
		url := "/" + activityImageDir + "/" + newFileName
		saved = append(saved, url)
		rows = append(rows, model.ActivityImage{
			ActivityID: activity.ID,
			ImgURL:     url,
			SortOrder:  maxOrder + i + 1,
		})
	}

	if err := database.DB.Create(&rows).Error; err != nil {
		for _, p := range saved { // DB พลาด → ไม่ให้มีไฟล์กำพร้าค้างไว้
			os.Remove("." + p)
		}
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": rows})
}

// DeleteActivityImage godoc
// @Summary  ลบรูปหนึ่งใบออกจากอัลบั้มกิจกรรม
// @Tags     activities
// @Produce  json
// @Security BearerAuth
// @Param    id      path int true "activity id"
// @Param    imageId path int true "activity image id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /activities/{id}/images/{imageId} [delete]
//
// ลบเฉพาะรูปในอัลบั้ม — รูปปก (Activity.ImgURL) เปลี่ยนผ่าน PUT /activities/{id} เหมือนเดิม
func DeleteActivityImage(c *gin.Context) {
	id := c.Param("id")
	imageID := c.Param("imageId")

	var img model.ActivityImage
	// ผูก activity_id ไว้ในเงื่อนไขด้วย กันลบรูปข้ามกิจกรรมด้วยการเดา id
	if err := database.DB.Where("id = ? AND activity_id = ?", imageID, id).First(&img).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบรูปนี้ในกิจกรรม"})
		return
	}

	if err := database.DB.Delete(&model.ActivityImage{}, img.ID).Error; err != nil {
		dbError(c, err)
		return
	}
	if img.ImgURL != "" {
		os.Remove("." + img.ImgURL) // best-effort
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
