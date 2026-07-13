package controllers

import (
	model "hospitalbackend/models"
	"net/http"

	"hospitalbackend/database"

	"github.com/gin-gonic/gin"
)

// GetAllMenu godoc
// @Summary  list เมนูทั้งหมด
// @Tags     menu
// @Produce  json
// @Success  200 {object} map[string]interface{}
// @Router   /menu [get]
func GetAllMenu(c *gin.Context) {
	var menus []model.Menu
	result := database.DB.Find(&menus)

	if result.Error != nil {
		dbError(c, result.Error)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    menus,
	})
}

// GetMenuByID godoc
// @Summary  ดูเมนูรายตัว
// @Tags     menu
// @Produce  json
// @Param    id path int true "menu id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /menu/{id} [get]
func GetMenuByID(c *gin.Context) {
	id := c.Param("id")

	var menu model.Menu
	result := database.DB.Where("id = ?", id).First(&menu)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "menu not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    menu,
	})
}

// AddNewMenu godoc
// @Summary  สร้างเมนูใหม่
// @Tags     menu
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    menu body model.Menu true "ข้อมูลเมนู"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Router   /menu [post]
func AddNewMenu(c *gin.Context) {
	var input model.Menu

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	menu := model.Menu{
		MenuName:    input.MenuName,
		Description: input.Description,
		Link:        input.Link,
		Icon:        input.Icon,
		Color:       input.Color,
	}

	result := database.DB.Create(&menu)
	if result.Error != nil {
		dbError(c, result.Error)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    menu,
	})
}

// UpdateMenu godoc
// @Summary  แก้ไขเมนู
// @Tags     menu
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id   path int true "menu id"
// @Param    menu body model.Menu true "ข้อมูลเมนู"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /menu/{id} [put]
func UpdateMenu(c *gin.Context) {
	id := c.Param("id")

	var menu model.Menu
	if err := database.DB.First(&menu, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบเมนู"})
		return
	}

	var input model.Menu
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	menu.MenuName = input.MenuName
	menu.Description = input.Description
	menu.Link = input.Link
	menu.Icon = input.Icon
	menu.Color = input.Color

	if err := database.DB.Save(&menu).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": menu})
}

// DeleteMenu godoc
// @Summary  ลบเมนู (เฉพาะ admin/employee)
// @Tags     menu
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "menu id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /menu/{id} [delete]
func DeleteMenu(c *gin.Context) {
	id := c.Param("id")

	var menu model.Menu
	if err := database.DB.First(&menu, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบเมนู"})
		return
	}

	if err := database.DB.Delete(&model.Menu{}, id).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
