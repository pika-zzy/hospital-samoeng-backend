package controllers

import (
	"encoding/json"
	models "hospitalbackend/models"
	"net/http"
	"strconv"
	"strings"

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
		Preload("Items.Children", func(db *gorm.DB) *gorm.DB {
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
// @Param    body body object{label=string,parent_id=int} true "ชื่อ item + ข้อแม่ (parent_id ไม่ใส่/null = ข้อชั้นบนสุด)"
// @Success  200 {object} map[string]interface{}
// @Router   /topics/{id}/items [post]
// validateItemParent ตรวจว่า parentID เอาไปเป็น "แม่" ของข้อใน topicID ได้ไหม
// คืนข้อความไทยถ้าไม่ผ่าน (คืน "" = ผ่าน)
//
// กติกา (ดูเหตุผลที่ models/ita_model.go):
//   - แม่ต้องมีจริง และอยู่ topic เดียวกัน
//   - แม่ต้องเป็นข้อชั้นบนสุด -> ลึกได้แค่ 2 ชั้น และวงวนเกิดไม่ได้
//   - ห้ามเป็นแม่ตัวเอง
func validateItemParent(parentID uint, topicID uint, selfID uint) string {
	if selfID != 0 && parentID == selfID {
		return "เลือกข้อตัวเองเป็นข้อแม่ไม่ได้"
	}

	var parent models.MoitItem
	if err := database.DB.First(&parent, parentID).Error; err != nil {
		return "ไม่พบข้อที่เลือกเป็นข้อแม่"
	}
	if parent.TopicID != topicID {
		return "ข้อแม่ต้องอยู่ในหัวข้อเดียวกัน"
	}
	if parent.ParentID != nil {
		return "ข้อแม่ที่เลือกเป็นข้อย่อยอยู่แล้ว — ระบบรองรับข้อย่อยชั้นเดียว"
	}
	return ""
}

// itemHasChildren — ใช้กันทั้งตอนย้ายข้อและตอนลบ
func itemHasChildren(id uint) (bool, error) {
	var n int64
	if err := database.DB.Model(&models.MoitItem{}).Where("parent_id = ?", id).Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func CreateItem(c *gin.Context) {
	topicID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var body struct {
		Label    string `json:"label" binding:"required"`
		ParentID *uint  `json:"parent_id"` // ไม่ส่ง/null = ข้อชั้นบนสุด
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	if body.ParentID != nil {
		if msg := validateItemParent(*body.ParentID, topicID, 0); msg != "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": msg})
			return
		}
	}

	item := models.MoitItem{
		TopicID:  topicID,
		Label:    body.Label,
		ParentID: body.ParentID,
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
// @Param    body body object{label=string,parent_id=int} true "ชื่อใหม่ + ข้อแม่ (ไม่ส่ง = ไม่แตะ, null = ขึ้นชั้นบนสุด)"
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

	// ParentID ต้องแยก 3 กรณีออกจากกัน:
	//   ไม่ส่ง field มาเลย -> ไม่แตะข้อแม่เดิม
	//   ส่ง null           -> ย้ายขึ้นเป็นข้อชั้นบนสุด
	//   ส่งเลข             -> ย้ายไปเป็นลูกของข้อนั้น
	//
	// FIX: เดิมใช้ **uint ซึ่งแยกไม่ออก — encoding/json เจอ null แล้ว "เซ็ต pointer
	// ตัวนอกเป็น nil" ผลคือ null กับไม่ส่ง field มาหน้าตาเหมือนกันเป๊ะ ทำให้สั่งย้าย
	// ข้อขึ้นชั้นบนสุดแล้วเงียบ (API ตอบ success แต่ DB ไม่เปลี่ยน)
	// RawMessage เก็บ bytes ดิบไว้ จึงบอกได้ว่า field มาไหมและมาเป็น null หรือเลข
	var body struct {
		Label    string          `json:"label" binding:"required"`
		ParentID json.RawMessage `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	if len(body.ParentID) > 0 {
		var newParent *uint
		if raw := strings.TrimSpace(string(body.ParentID)); raw != "null" {
			var pid uint
			if err := json.Unmarshal(body.ParentID, &pid); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
				return
			}
			newParent = &pid
		}

		if newParent != nil {
			// ข้อที่มีลูกอยู่แล้ว ย้ายลงไปเป็นลูกของใครไม่ได้ (จะกลายเป็น 3 ชั้น)
			has, err := itemHasChildren(item.ID)
			if err != nil {
				dbError(c, err)
				return
			}
			if has {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "ข้อนี้มีข้อย่อยอยู่ ย้ายไปเป็นข้อย่อยของข้ออื่นไม่ได้",
				})
				return
			}
			if msg := validateItemParent(*newParent, item.TopicID, item.ID); msg != "" {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": msg})
				return
			}
		}
		item.ParentID = newParent
	}

	item.Label = body.Label

	// ใช้ Select ระบุคอลัมน์ตรง ๆ เพราะ Save ของ GORM ข้าม field ที่เป็น zero value
	// (ParentID = nil จะไม่ถูกเขียนลง DB ถ้าไม่บอก)
	if err := database.DB.Model(&item).Select("label", "parent_id").Updates(map[string]any{
		"label":     item.Label,
		"parent_id": item.ParentID,
	}).Error; err != nil {
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

	// มีข้อย่อยอยู่ -> ไม่ลบให้ (FK เป็น CASCADE ถ้าปล่อยผ่านจะลากข้อย่อย
	// และไฟล์ที่แขวนอยู่ใต้ข้อย่อยหายไปด้วยเงียบ ๆ)
	has, err := itemHasChildren(id)
	if err != nil {
		dbError(c, err)
		return
	}
	if has {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ข้อนี้มีข้อย่อยอยู่ ลบข้อย่อยให้หมดก่อนถึงจะลบได้",
		})
		return
	}

	if err := database.DB.Delete(&models.MoitItem{}, id).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}