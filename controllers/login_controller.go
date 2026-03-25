package controllers

import (
	"hospitalbackend/database"
	model "hospitalbackend/models"
	"hospitalbackend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "user not found",
		})
		return
	}
	//ตรวจรหัส
	if member.Password != inputUser.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "wrong password",
		})
		return
	}

	token, err := utils.GenerateToken(member.ID, member.Role)

	if err != nil {
		c.JSON(500, gin.H{"message": "token error"})
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
