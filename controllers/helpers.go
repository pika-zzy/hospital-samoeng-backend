package controllers

import (
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// dbError — log error ฝั่ง server (พร้อม method+path ให้ตามรอยได้) แล้วตอบ 500
// ด้วยข้อความกลาง ๆ — ห้ามส่ง raw DB error กลับ client เพราะเผยรายละเอียด schema/query
func dbError(c *gin.Context, err error) {
	log.Printf("internal error: %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง"})
}

// ==================== File upload validation (CRIT-03) ====================
//
// เพดานขนาดไฟล์ต่อไฟล์ — ปรับที่นี่ที่เดียว มีผลกับทุก upload handler
// (ค่า default: รูป 5MB, PDF 25MB — เอกสารสแกน ITA ส่วนใหญ่ต่ำกว่านี้)
const (
	MaxImageBytes int64 = 5 << 20  // 5 MB — .jpg/.jpeg/.png
	MaxPDFBytes   int64 = 25 << 20 // 25 MB — .pdf (ITA/ข่าว)

	// maxUploadMemory — RAM buffer ตอน parse multipart; ส่วนเกิน spill ลง temp disk
	maxUploadMemory int64 = 8 << 20 // 8 MB
	// maxFormOverhead — เผื่อ text field อื่น ๆ ในคำขอ multipart เดียวกัน
	maxFormOverhead int64 = 1 << 20 // 1 MB
)

// ชนิดไฟล์ที่อนุญาต (นามสกุลตัวพิมพ์เล็ก — เทียบหลัง strings.ToLower)
var (
	allowedImageExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	allowedPDFExt   = map[string]bool{".pdf": true}

	// ITA รับได้ทั้ง PDF และรูป — MOIT บางข้อขอหลักฐานเป็นภาพถ่าย/อินโฟกราฟิก
	// (เช่น MOIT20 "ให้จัดทำอินโฟกราฟิกการเข้าร่วมกิจกรรมวันต่อต้านคอร์รัปชัน")
	allowedITAExt = map[string]bool{".pdf": true, ".jpg": true, ".jpeg": true, ".png": true}
)

// itaMaxBytes — เพดานขนาดตามชนิดไฟล์ที่ ITA รับ (รูปเล็กกว่า PDF)
func itaMaxBytes(filename string) int64 {
	if strings.ToLower(filepath.Ext(filename)) == ".pdf" {
		return MaxPDFBytes
	}
	return MaxImageBytes
}

// enforceUploadLimit จำกัดขนาด body รวมของคำขอ (C1) แล้ว parse multipart ทันที
// ต้องเรียกเป็นด่านแรกสุดของ upload handler ก่อนอ่าน field/ไฟล์ใด ๆ —
// ถ้า body เกินเพดานจะ fail ตรงนี้ (คืน error) แทนที่จะถูกมองเป็น "ไม่มีไฟล์" เงียบ ๆ
func enforceUploadLimit(c *gin.Context, maxBodyBytes int64) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
	return c.Request.ParseMultipartForm(maxUploadMemory)
}

// uploadLimitError ตอบ client ตามชนิด error จาก enforceUploadLimit:
// body เกินเพดาน → 413, อื่น ๆ (เช่น content-type ไม่ใช่ multipart) → 400
func uploadLimitError(c *gin.Context, err error) {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"success": false, "message": "ไฟล์หรือคำขอมีขนาดใหญ่เกินกำหนด"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
}

// validateUpload ตรวจไฟล์ตามลำดับ: (2) ขนาด → (3) นามสกุล case-insensitive
// → (4) sniff MIME จริง → (5) เทียบ MIME กับนามสกุล
// คืนนามสกุลตัวพิมพ์เล็กที่ normalize แล้ว (ให้ caller เอาไปตั้งชื่อไฟล์) หรือ error ข้อความไทย
func validateUpload(file *multipart.FileHeader, allowedExt map[string]bool, maxBytes int64) (string, error) {
	// (2) ขนาดไฟล์
	if file.Size > maxBytes {
		return "", errors.New("ไฟล์มีขนาดใหญ่เกินกำหนด")
	}

	// (3) นามสกุล — case-insensitive กัน .JPG/.PDF จากมือถือถูก reject (L4)
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		return "", errors.New("ชนิดไฟล์ไม่ถูกต้อง")
	}

	// (4) sniff MIME จริงจาก 512 ไบต์แรก (http.DetectContentType ใช้แค่นี้)
	f, err := file.Open()
	if err != nil {
		return "", errors.New("เปิดไฟล์ไม่สำเร็จ")
	}
	defer f.Close()

	head, err := io.ReadAll(io.LimitReader(f, 512))
	if err != nil {
		return "", errors.New("อ่านไฟล์ไม่สำเร็จ")
	}

	// (5) เทียบ MIME กับนามสกุล — กันไฟล์ปลอมชนิด (H1)
	if !mimeMatchesExt(ext, http.DetectContentType(head)) {
		return "", errors.New("เนื้อหาไฟล์ไม่ตรงกับนามสกุล")
	}

	return ext, nil
}

// mimeMatchesExt เทียบ MIME ที่ sniff ได้กับนามสกุลที่อ้าง
func mimeMatchesExt(ext, detected string) bool {
	switch ext {
	case ".jpg", ".jpeg":
		return detected == "image/jpeg"
	case ".png":
		return detected == "image/png"
	case ".pdf":
		return detected == "application/pdf"
	default:
		return false
	}
}
