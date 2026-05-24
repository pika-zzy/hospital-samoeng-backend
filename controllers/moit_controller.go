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

// GET /ita/years/:id/moit
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

		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": moits})
}


// POST /ita/years/:id/moit
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	moit := models.MoitCategory{
		Name:        body.Name,
		Description: body.Description,
		YearID:      yearID,
	}

	if err := database.DB.Create(&moit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": moit})
}


// PUT /moit/:id
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	if body.Name != "" {
		moit.Name = body.Name
	}
	if body.Description != nil {
		moit.Description = body.Description
	}

	if err := database.DB.Save(&moit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": moit})
}


// DELETE /moit/:id
func DeleteMoit(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := database.DB.Delete(&models.MoitCategory{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}


// ==================== TOPIC ====================

// GET /moit/:id/topics
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

		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": topics})
}


// POST /moit/:id/topics
func CreateTopic(c *gin.Context) {
	moitID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var body struct {
		Label string `json:"label" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	topic := models.MoitTopic{
		MoitID: moitID,
		Label:  body.Label,
	}

	if err := database.DB.Create(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": topic})
}


// PUT /topics/:id
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	topic.Label = body.Label

	if err := database.DB.Save(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": topic})
}


// DELETE /topics/:id
func DeleteTopic(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := database.DB.Delete(&models.MoitTopic{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}


// ==================== ITEM ====================

// POST /topics/:id/items
func CreateItem(c *gin.Context) {
	topicID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var body struct {
		Label string `json:"label" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	item := models.MoitItem{
		TopicID: topicID,
		Label:   body.Label,
	}

	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}


// PUT /items/:id
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	item.Label = body.Label

	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}


// DELETE /items/:id
func DeleteItem(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := database.DB.Delete(&models.MoitItem{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}