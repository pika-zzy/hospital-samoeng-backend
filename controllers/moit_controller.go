package controllers

import (
	models "hospitalbackend/models"
	"net/http"
	"strconv"

	"hospitalbackend/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


// ==================== UTIL ====================

func parseUintParam(c *gin.Context, key string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": key + " ไม่ถูกต้อง"})
		return 0, false
	}
	return uint(id), true
}


// ==================== MOIT ====================

// GetMoitByYear godoc
// @Summary  list MOIT ของปีที่เลือก
// @Tags     moit
// @Produce  json
// @Param    id path int true "year id"
// @Success  200 {object} map[string]interface{}
// @Router   /ita/years/{id}/moit [get]
func GetMoitByYear(c *gin.Context) {
	yearID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var moits []models.MoitCategory
	if err := database.DB.
		Where("year_id = ?", yearID).
		Order("id asc").
		Find(&moits).Error; err != nil {

		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": moits})
}


// CreateMoit godoc
// @Summary  สร้าง MOIT ในปีที่เลือก (เฉพาะ admin)
// @Tags     moit
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id   path int true "year id"
// @Param    body body object{name=string,description=string} true "ชื่อ MOIT"
// @Success  200 {object} map[string]interface{}
// @Router   /ita/years/{id}/moit [post]
func CreateMoit(c *gin.Context) {
	yearID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var body struct {
		Name        string `json:"name" binding:"required"`
		Description *string `json:"description"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	moit := models.MoitCategory{
		Name:        body.Name,
		Description: body.Description,
		YearID:      yearID,
	}

	if err := database.DB.Create(&moit).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": moit})
}


// UpdateMoit godoc
// @Summary  แก้ไข MOIT (เฉพาะ admin)
// @Tags     moit
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id   path int true "moit id"
// @Param    body body object{name=string,description=string} true "ฟิลด์ที่จะแก้"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /moit/{id} [put]
func UpdateMoit(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var moit models.MoitCategory
	if err := database.DB.First(&moit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบ MOIT"})
		return
	}

	var body struct {
		Name        string `json:"name"`
		Description *string `json:"description"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	if body.Name != "" {
		moit.Name = body.Name
	}
	if body.Description != nil {
		moit.Description = body.Description
	}

	if err := database.DB.Save(&moit).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": moit})
}


// DeleteMoit godoc
// @Summary  ลบ MOIT (เฉพาะ admin, ลูก topic/item/ไฟล์หายตาม CASCADE)
// @Tags     moit
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "moit id"
// @Success  200 {object} map[string]interface{}
// @Router   /moit/{id} [delete]
func DeleteMoit(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := database.DB.Delete(&models.MoitCategory{}, id).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}


// ==================== TOPIC ====================

// GetMoitTopics godoc
// @Summary  list topic (พร้อม items) ของ MOIT
// @Tags     moit
// @Produce  json
// @Param    id path int true "moit id"
// @Success  200 {object} map[string]interface{}
// @Router   /moit/{id}/topics [get]
func GetMoitTopics(c *gin.Context) {
	moitID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var topics []models.MoitTopic
	if err := database.DB.
		Where("moit_id = ?", moitID).
		Order("label asc").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("label asc")
		}).
		Find(&topics).Error; err != nil {

		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": topics})
}


// CreateTopic godoc
// @Summary  สร้าง topic ใน MOIT (เฉพาะ admin)
// @Tags     moit
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id   path int true "moit id"
// @Param    body body object{label=string} true "ชื่อ topic"
// @Success  200 {object} map[string]interface{}
// @Router   /moit/{id}/topics [post]
func CreateTopic(c *gin.Context) {
	moitID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var body struct {
		Label string `json:"label" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	topic := models.MoitTopic{
		MoitID: moitID,
		Label:  body.Label,
	}

	if err := database.DB.Create(&topic).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": topic})
}


// UpdateTopic godoc
// @Summary  แก้ไข topic (เฉพาะ admin)
// @Tags     moit
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id   path int true "topic id"
// @Param    body body object{label=string} true "ชื่อใหม่"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /topics/{id} [put]
func UpdateTopic(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var topic models.MoitTopic
	if err := database.DB.First(&topic, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบ Topic"})
		return
	}

	var body struct {
		Label string `json:"label" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	topic.Label = body.Label

	if err := database.DB.Save(&topic).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": topic})
}


// DeleteTopic godoc
// @Summary  ลบ topic (เฉพาะ admin)
// @Tags     moit
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "topic id"
// @Success  200 {object} map[string]interface{}
// @Router   /topics/{id} [delete]
func DeleteTopic(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := database.DB.Delete(&models.MoitTopic{}, id).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}


// ==================== ITEM ====================

// CreateItem godoc
// @Summary  สร้าง item ใน topic (เฉพาะ admin)
// @Tags     moit
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id   path int true "topic id"
// @Param    body body object{label=string} true "ชื่อ item"
// @Success  200 {object} map[string]interface{}
// @Router   /topics/{id}/items [post]
func CreateItem(c *gin.Context) {
	topicID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var body struct {
		Label string `json:"label" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	item := models.MoitItem{
		TopicID: topicID,
		Label:   body.Label,
	}

	if err := database.DB.Create(&item).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}


// UpdateItem godoc
// @Summary  แก้ไข item (เฉพาะ admin)
// @Tags     moit
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id   path int true "item id"
// @Param    body body object{label=string} true "ชื่อใหม่"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /items/{id} [put]
func UpdateItem(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var item models.MoitItem
	if err := database.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบ Item"})
		return
	}

	var body struct {
		Label string `json:"label" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	item.Label = body.Label

	if err := database.DB.Save(&item).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}


// DeleteItem godoc
// @Summary  ลบ item (เฉพาะ admin)
// @Tags     moit
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "item id"
// @Success  200 {object} map[string]interface{}
// @Router   /items/{id} [delete]
func DeleteItem(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := database.DB.Delete(&models.MoitItem{}, id).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}