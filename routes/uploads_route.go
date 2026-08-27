package routes

import (
	"net/http"
	"path"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

// UploadsRoute เสิร์ฟไฟล์ใน ./uploads แบบ public (ไม่มี auth) เหมือน r.Static เดิมทุกอย่าง
// เพิ่มมาอย่างเดียวคือ query `?download=<ชื่อไฟล์>`:
//
//	ไม่มี download            -> เปิดดูในเบราว์เซอร์ (PDF/รูปแสดง inline เหมือนเดิม)
//	?download=1               -> บังคับดาวน์โหลด ใช้ชื่อไฟล์จริงบนดิสก์ (timestamp)
//	?download=ชื่อประกาศ.pdf   -> บังคับดาวน์โหลด ใช้ชื่อที่ส่งมา (ผ่าน sanitize แล้ว)
//
// ทำไมต้องมี: ไฟล์ถูกเสิร์ฟจากคนละ origin กับหน้าเว็บ (8080 vs 5173 และคนละ subdomain
// บน production) attribute `download` ของ <a> จึงถูกเบราว์เซอร์เมินทั้งหมด
// การบังคับดาวน์โหลดข้าม origin ทำได้ทางเดียวคือให้ฝั่ง server ส่ง Content-Disposition
func UploadsRoute(r *gin.Engine) {
	fileServer := http.StripPrefix("/uploads", http.FileServer(http.Dir("./uploads")))

	r.GET("/uploads/*filepath", func(c *gin.Context) {
		if name := c.Query("download"); name != "" {
			filename := safeDownloadName(name, path.Base(c.Param("filepath")))
			// RFC 6266: ส่งทั้ง filename (ASCII fallback) และ filename* (UTF-8)
			// เพื่อให้ชื่อไทยไม่กลายเป็นตัวขยะบนเบราว์เซอร์เก่า
			c.Header("Content-Disposition",
				`attachment; filename="`+asciiFallback(filename)+`"; filename*=UTF-8''`+urlEscape(filename))
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

// safeDownloadName แปลงชื่อที่ client ขอมาให้ปลอดภัยพอจะใส่ใน header
// - "1" (ค่าเริ่มต้นจาก frontend) = ใช้ชื่อไฟล์จริงบนดิสก์
// - ตัด path separator, quote, และอักขระควบคุมทั้งหมด (กัน header injection)
// - คงนามสกุลจริงของไฟล์บนดิสก์เสมอ ไม่เชื่อส่วนขยายที่ client ส่งมา
// - จำกัดความยาวกันชื่อยาวเกินจนบาง client ตัดทิ้งทั้ง header
func safeDownloadName(requested, actual string) string {
	if requested == "1" {
		return actual
	}

	ext := path.Ext(actual)
	base := strings.TrimSuffix(requested, path.Ext(requested))

	var b strings.Builder
	for _, r := range base {
		switch {
		case r == '/' || r == '\\' || r == '"' || r == ';':
			continue
		case unicode.IsControl(r):
			continue
		default:
			b.WriteRune(r)
		}
	}

	clean := strings.TrimSpace(b.String())
	// นับเป็น rune ไม่ใช่ byte — ตัดกลาง rune ภาษาไทยจะได้ชื่อพัง
	if runes := []rune(clean); len(runes) > 80 {
		clean = strings.TrimSpace(string(runes[:80]))
	}
	if clean == "" {
		return actual
	}
	return clean + ext
}

// asciiFallback ทิ้งอักขระนอก ASCII ออกสำหรับ client เก่าที่อ่าน filename* ไม่เป็น
// ถ้าตัดแล้วไม่เหลืออะไร (ชื่อไทยล้วน) ให้ใช้ "download" + นามสกุลแทน
func asciiFallback(name string) string {
	var b strings.Builder
	for _, r := range name {
		if r < unicode.MaxASCII && r > 31 && r != '"' && r != '\\' {
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" || out == path.Ext(name) {
		return "download" + path.Ext(name)
	}
	return out
}

// urlEscape เข้ารหัสทุก byte ที่ไม่ใช่ unreserved ตาม RFC 3986
// (ใช้แทน url.QueryEscape ที่แปลง space เป็น '+' ซึ่งผิดสำหรับ filename*)
func urlEscape(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for _, c := range []byte(s) {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '~' {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&0x0f])
	}
	return b.String()
}
