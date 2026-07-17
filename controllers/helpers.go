package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// dbError — log error ฝั่ง server (พร้อม method+path ให้ตามรอยได้) แล้วตอบ 500
// ด้วยข้อความกลาง ๆ — ห้ามส่ง raw DB error กลับ client เพราะเผยรายละเอียด schema/query
func dbError(c *gin.Context, err error) {
	log.Printf("internal error: %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง"})
}
