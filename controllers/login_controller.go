package controllers

import (
	"errors"
	"hospitalbackend/database"
	model "hospitalbackend/models"
	"hospitalbackend/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Login godoc
// @Summary  ล็อกอิน รับ JWT token
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body body object{username=string,password=string} true "ข้อมูลล็อกอิน"
// @Success  200 {object} map[string]interface{} "token + ข้อมูล user"
// @Failure  401 {object} map[string]interface{} "user ไม่พบหรือรหัสผิด"
// @Failure  403 {object} map[string]interface{} "บัญชีถูกปิดใช้งาน"
// @Router   /login [post]
func Login(c *gin.Context) {
	var inputUser model.User
	// รับ JSON จาก frontend
	if err := c.ShouldBindJSON(&inputUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid input",
		})
		return
	}

	var member model.User
	//หา user จาก username

	if err := database.DB.Where("username = ?", inputUser.Username).First(&member).Error; err != nil {
		// FIX: เพิ่ม success:false ให้ครบตาม envelope ที่ frontend เช็ค
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "user not found or wrong password",
		})
		return
	}

	// เช็คสถานะว่าถูกปิดใช้งานไหม
	if !member.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "บัญชีนี้ถูกปิดใช้งาน",
		})
		return
	}
	//ตรวจรหัส
	if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(inputUser.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "user not found or wrong password",
		})
		return
	}

	token, err := utils.GenerateToken(member.ID, member.Role)

	if err != nil {
		// FIX: log error ฝั่ง server + ใช้ constant แทนเลข 500 ดิบ
		log.Println("GenerateToken error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "token error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "login success",
		"token":   token,
		"user": gin.H{
			"id":   member.ID,
			"name": member.Username,
			"role": member.Role,
		},
	})

}

// GetAllUsers godoc
// @Summary  list user ทั้งหมด (ไม่ส่ง password กลับ)
// @Tags     users
// @Produce  json
// @Security BearerAuth
// @Success  200 {object} map[string]interface{}
// @Router   /users [get]
func GetAllUsers(c *gin.Context) {
	var users []model.User
	if err := database.DB.Find(&users).Error; err != nil {
		// FIX: ไม่ส่ง raw DB error กลับหา client — log ฝั่ง server แทน
		log.Println("GetAllUsers error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง"})
		return
	}

	data := make([]gin.H, 0, len(users))
	for _, u := range users {
		data = append(data, gin.H{
			"id":        u.ID,
			"username":  u.Username,
			"role":      u.Role,
			"is_active": u.IsActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// UpdateUserStatus godoc
// @Summary  เปิด/ปิดการใช้งานบัญชี (ห้ามปิดบัญชีตัวเอง)
// @Tags     users
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id   path int true "user id"
// @Param    body body object{is_active=bool} true "สถานะใหม่"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{} "ข้อมูลผิด/ปิดบัญชีตัวเอง"
// @Failure  404 {object} map[string]interface{}
// @Router   /users/{id}/status [patch]
func UpdateUserStatus(c *gin.Context) {
	id := c.Param("id")

	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบผู้ใช้งาน"})
		return
	}

	var body struct {
		IsActive *bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.IsActive == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	// FIX: กัน admin ปิดใช้งานบัญชีตัวเอง — ปิดแล้วจะ login กลับมาแก้ไม่ได้อีก
	if userID, exists := c.Get("userID"); exists && userID == user.ID && !*body.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ไม่สามารถปิดใช้งานบัญชีของตัวเองได้"})
		return
	}

	if err := database.DB.Model(&user).Update("is_active", *body.IsActive).Error; err != nil {
		log.Println("UpdateUserStatus error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"role":      user.Role,
			"is_active": *body.IsActive,
		},
	})
}

// GetUserByID godoc
// @Summary  ดูข้อมูล user รายคน (ไม่ส่ง password กลับ)
// @Tags     users
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "user id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /users/{id} [get]
func GetUserByID(c *gin.Context) {
	id := c.Param("id")

	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบผู้ใช้งาน"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"role":      user.Role,
			"is_active": user.IsActive,
		},
	})
}

// AddUser godoc
// @Summary  สร้าง user ใหม่ (เฉพาะ admin)
// @Tags     users
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body object{username=string,password=string,role=string} true "role: employee หรือ admin"
// @Success  201 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Failure  409 {object} map[string]interface{} "username ซ้ำ"
// @Router   /users [post]
func AddUser(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Username == "" || body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "กรุณากรอก username และ password"})
		return
	}

	if body.Role != "employee" && body.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "role ต้องเป็น employee หรือ admin"})
		return
	}

	// เช็ค username ซ้ำ
	var existing model.User
	err := database.DB.Where("username = ?", body.Username).First(&existing).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "username นี้ถูกใช้งานแล้ว"})
		return
	}
	// FIX: แยกเคส "ไม่เจอ" (ปกติ ไปต่อได้) ออกจาก DB error จริง — เดิมถ้า DB error จะหลุดไป Create ต่อ
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("AddUser check duplicate error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "เกิดข้อผิดพลาดในการสร้างบัญชี"})
		return
	}

	// NOTE: ต้อง set IsActive: true ตรงนี้เอง — ปล่อยเป็น zero value (false) แล้วหวังพึ่ง
	// gorm default:true ไม่ได้ เพราะพฤติกรรม GORM กับ bool zero value ไม่แน่นอน
	user := model.User{
		Username: body.Username,
		Password: string(hashed),
		Role:     body.Role,
		IsActive: true,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		log.Println("AddUser error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"role":      user.Role,
			"is_active": user.IsActive,
		},
	})
}

// DeleteUser godoc
// @Summary  ลบ user (เฉพาะ admin, ห้ามลบบัญชีตัวเอง)
// @Tags     users
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "user id"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{} "ลบบัญชีตัวเอง"
// @Failure  404 {object} map[string]interface{}
// @Router   /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบผู้ใช้งาน"})
		return
	}

	// กันเคส admin ลบบัญชีตัวเอง (userID มาจาก AuthMiddleware)
	if userID, exists := c.Get("userID"); exists && userID == user.ID {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ไม่สามารถลบบัญชีของตัวเองได้"})
		return
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		log.Println("DeleteUser error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "ลบผู้ใช้งานสำเร็จ"})
}