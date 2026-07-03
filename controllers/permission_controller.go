package controllers

import (
	"hospitalbackend/database"
	model "hospitalbackend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// role ที่ระบบรองรับตอนนี้ — ตรงกับค่าใน User.Role
func isValidRole(role string) bool {
	return role == "employee" || role == "admin"
}

// GetRolePermissions godoc
// @Summary  รายการ menu key ที่ role นี้มองเห็น
// @Description ตอบ []string ของ NAV.to เช่น ["/admin/dashboard"] — list ว่าง = ยังไม่เคยตั้งค่า = เห็นทุกเมนู
// @Tags     roles
// @Produce  json
// @Security BearerAuth
// @Param    role path string true "employee หรือ admin"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{} "role ไม่ถูกต้อง"
// @Router   /roles/{role}/permissions [get]
func GetRolePermissions(c *gin.Context) {
	role := c.Param("role")
	if !isValidRole(role) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "role ไม่ถูกต้อง"})
		return
	}

	var perms []model.RolePermission
	if err := database.DB.Where("role = ?", role).Find(&perms).Error; err != nil {
		dbError(c, err)
		return
	}

	menus := make([]string, 0, len(perms))
	for _, p := range perms {
		menus = append(menus, p.MenuKey)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": menus})
}

// UpdateRolePermissions godoc
// @Summary  ตั้งสิทธิ์เมนูของ role ทั้งชุด (เฉพาะ admin)
// @Description แทนที่สิทธิ์เดิมทั้งหมดใน transaction เดียว — ห้ามแก้ role admin (เห็นทุกเมนูเสมอ)
// @Tags     roles
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    role path string true "employee เท่านั้น (admin แก้ไม่ได้)"
// @Param    body body object{menus=[]string} true "list ของ menu key (NAV.to)"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Router   /roles/{role}/permissions [put]
func UpdateRolePermissions(c *gin.Context) {
	role := c.Param("role")
	if !isValidRole(role) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "role ไม่ถูกต้อง"})
		return
	}

	// admin เห็นทุกเมนูเสมอ — ไม่ให้แก้ กัน admin เผลอล็อกตัวเองออกจากระบบ
	if role == "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ไม่สามารถแก้ไขสิทธิ์ของ admin ได้"})
		return
	}

	var body struct {
		Menus []string `json:"menus"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Menus == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	tx := database.DB.Begin()
	if err := tx.Where("role = ?", role).Delete(&model.RolePermission{}).Error; err != nil {
		tx.Rollback()
		dbError(c, err)
		return
	}
	for _, menu := range body.Menus {
		if err := tx.Create(&model.RolePermission{Role: role, MenuKey: menu}).Error; err != nil {
			tx.Rollback()
			dbError(c, err)
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": body.Menus})
}
