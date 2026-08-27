package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hospitalbackend/database"
	model "hospitalbackend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ไฟล์แนบของหน้าเนื้อหา — รับได้ทั้ง PDF และรูป (เว็บเก่ามีทั้งสองแบบ)
const contentFileDir = "uploads/file/content"

// maxContentFilesPerRequest — กันคำขอเดียวยัดไฟล์ไม่จำกัด ไม่ใช่เพดานต่อหัวข้อ
// (หน้าเนื้อหาไม่มีเพดานรวมแบบอัลบั้มกิจกรรม เพราะเอกสารสะสมเพิ่มทุกปี)
const maxContentFilesPerRequest = 20

// ---------- helper ----------

// findSectionBySlug — ทุก endpoint ของ section อ้างด้วย slug ไม่ใช่ id
// เพื่อให้ URL หน้าเว็บ (/about/<slug>) กับ API ใช้กุญแจตัวเดียวกัน
func findSectionBySlug(slug string) (*model.ContentSection, error) {
	var s model.ContentSection
	if err := database.DB.Where("slug = ?", slug).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// removeFilesOnDisk — best-effort ลบไฟล์จริงตาม URL ที่เก็บใน DB
// เช็ค prefix ก่อนลบ กัน path แปลกปลอมใน DB พาไปลบไฟล์นอกโฟลเดอร์ของตัวเอง
func removeFilesOnDisk(urls []string) {
	for _, u := range urls {
		if strings.HasPrefix(u, "/"+contentFileDir+"/") {
			os.Remove("." + u)
		}
	}
}

// normalizeSlug — บังคับ slug ให้เหลือแค่ ascii ตัวเล็ก/ตัวเลข/ขีด
// (slug ไปเป็นส่วนหนึ่งของ URL ปล่อยอักขระอื่นแล้วลิงก์เพี้ยนข้ามระบบ)
func normalizeSlug(raw string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(raw)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == ' ':
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// validBuddhistYear — กันปีพิมพ์ผิด (ใส่ ค.ศ. หรือ 2 หลัก) แล้วปุ่มสลับปีเพี้ยน
func validBuddhistYear(y int) bool { return y >= 2500 && y <= 2700 }

// ---------- public ----------

// GetContentSections godoc
// @Summary  รายชื่อหน้าเนื้อหาทั้งหมด (ไม่รวมหัวข้อ/ไฟล์ข้างใน)
// @Tags     content
// @Produce  json
// @Success  200 {object} map[string]interface{}
// @Router   /content/sections [get]
func GetContentSections(c *gin.Context) {
	var sections []model.ContentSection
	if err := database.DB.Order("sort_order ASC, id ASC").Find(&sections).Error; err != nil {
		dbError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": sections})
}

// GetContentSection godoc
// @Summary  หน้าเนื้อหา 1 หน้า พร้อมหัวข้อและไฟล์ทั้งหมด
// @Tags     content
// @Produce  json
// @Param    slug path string true "slug ของหน้า เช่น drug-safety"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /content/sections/{slug} [get]
//
// หัวข้อเรียงปีใหม่ก่อน (year DESC) แล้วค่อย sort_order — หัวข้อที่ไม่ผูกปี
// (year = NULL) MySQL จัดไว้ท้ายสุดของ DESC ซึ่งตรงกับที่ต้องการ: หน้าที่ไม่มีปี
// ทั้งหน้าก็เรียงด้วย sort_order ล้วน ๆ อยู่แล้ว
func GetContentSection(c *gin.Context) {
	section, err := findSectionBySlug(c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบหน้านี้"})
		return
	}

	if err := database.DB.
		Preload("Groups", func(db *gorm.DB) *gorm.DB {
			return db.Order("year DESC, sort_order ASC, id ASC")
		}).
		Preload("Groups.Files", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		First(section, section.ID).Error; err != nil {
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": section})
}

// ---------- section (admin) ----------

// CreateContentSection godoc
// @Summary  เพิ่มหน้าเนื้อหาใหม่
// @Tags     content
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body map[string]interface{} true "slug, title, description, sort_order"
// @Success  201 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Router   /content/sections [post]
//
// หน้าใหม่จะยังไม่โผล่ในเมนู "เกี่ยวกับเรา" จนกว่าจะเพิ่มบรรทัดใน
// frontend src/interface/menu.ts (เมนูหลักยังเป็นโค้ด ไม่ได้อ่านจาก DB)
// แต่เข้าถึงได้ทันทีทาง /about/<slug>
func CreateContentSection(c *gin.Context) {
	var body struct {
		Slug        string `json:"slug" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	slug := normalizeSlug(body.Slug)
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "slug ต้องมีตัวอักษรอังกฤษหรือตัวเลขอย่างน้อย 1 ตัว"})
		return
	}

	var dup int64
	if err := database.DB.Model(&model.ContentSection{}).Where("slug = ?", slug).Count(&dup).Error; err != nil {
		dbError(c, err)
		return
	}
	if dup > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "slug นี้ถูกใช้แล้ว"})
		return
	}

	section := model.ContentSection{
		Slug:        slug,
		Title:       strings.TrimSpace(body.Title),
		Description: strings.TrimSpace(body.Description),
		SortOrder:   body.SortOrder,
	}
	if err := database.DB.Create(&section).Error; err != nil {
		dbError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": section})
}

// UpdateContentSection godoc
// @Summary  แก้ชื่อ/คำอธิบาย/ลำดับของหน้าเนื้อหา
// @Tags     content
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    slug path string true "slug ของหน้า"
// @Param    body body map[string]interface{} true "title, description, sort_order"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /content/sections/{slug} [put]
//
// ตั้งใจไม่ให้แก้ slug — slug อยู่ใน URL ที่คนบันทึก/ส่งต่อกันไปแล้ว
func UpdateContentSection(c *gin.Context) {
	section, err := findSectionBySlug(c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบหน้านี้"})
		return
	}

	var body struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	section.Title = strings.TrimSpace(body.Title)
	section.Description = strings.TrimSpace(body.Description)
	section.SortOrder = body.SortOrder

	if err := database.DB.Save(section).Error; err != nil {
		dbError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": section})
}

// DeleteContentSection godoc
// @Summary  ลบหน้าเนื้อหา พร้อมหัวข้อและไฟล์ทั้งหมดข้างใน
// @Tags     content
// @Produce  json
// @Security BearerAuth
// @Param    slug path string true "slug ของหน้า"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /content/sections/{slug} [delete]
//
// model ใช้ gorm.Model = soft delete → FK ON DELETE CASCADE **ไม่ทำงาน**
// ในเส้นทางปกติ ต้องไล่ลบลูกหลานเองทั้งแถวใน DB และไฟล์บนดิสก์
// (บทเรียนเดียวกับ DeleteActivity — ปล่อยไว้จะเหลือแถวกำพร้ากับไฟล์ขยะ)
func DeleteContentSection(c *gin.Context) {
	section, err := findSectionBySlug(c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบหน้านี้"})
		return
	}

	var groupIDs []uint
	if err := database.DB.Model(&model.ContentGroup{}).
		Where("section_id = ?", section.ID).Pluck("id", &groupIDs).Error; err != nil {
		dbError(c, err)
		return
	}

	var urls []string
	if len(groupIDs) > 0 {
		if err := database.DB.Model(&model.ContentFile{}).
			Where("group_id IN ?", groupIDs).Pluck("file_url", &urls).Error; err != nil {
			dbError(c, err)
			return
		}
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if len(groupIDs) > 0 {
			if err := tx.Where("group_id IN ?", groupIDs).Delete(&model.ContentFile{}).Error; err != nil {
				return err
			}
			if err := tx.Where("section_id = ?", section.ID).Delete(&model.ContentGroup{}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&model.ContentSection{}, section.ID).Error
	})
	if err != nil {
		dbError(c, err)
		return
	}

	removeFilesOnDisk(urls) // ลบไฟล์หลัง DB สำเร็จเท่านั้น — DB พลาดแล้วไฟล์ยังอยู่ กู้ได้
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ---------- group (admin) ----------

// CreateContentGroup godoc
// @Summary  เพิ่มหัวข้อในหน้าเนื้อหา
// @Tags     content
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    slug path string true "slug ของหน้า"
// @Param    body body map[string]interface{} true "title, body, year (พ.ศ. หรือ null), sort_order"
// @Success  201 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /content/sections/{slug}/groups [post]
func CreateContentGroup(c *gin.Context) {
	section, err := findSectionBySlug(c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบหน้านี้"})
		return
	}

	var body struct {
		Title     string `json:"title" binding:"required"`
		Body      string `json:"body"`
		Year      *int   `json:"year"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}
	if body.Year != nil && !validBuddhistYear(*body.Year) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ปีต้องเป็น พ.ศ. เช่น 2569"})
		return
	}

	group := model.ContentGroup{
		SectionID: section.ID,
		Title:     strings.TrimSpace(body.Title),
		Body:      body.Body,
		Year:      body.Year,
		SortOrder: body.SortOrder,
	}
	if err := database.DB.Create(&group).Error; err != nil {
		dbError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": group})
}

// UpdateContentGroup godoc
// @Summary  แก้หัวข้อในหน้าเนื้อหา
// @Tags     content
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id   path int true "group id"
// @Param    body body map[string]interface{} true "title, body, year, sort_order"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /content/groups/{id} [put]
//
// year เป็น tri-state เหมือน parent_id ของ MoitItem จึงต้องรับเป็น json.RawMessage
// ไม่ส่ง field มา = ไม่แตะปีเดิม · null = ถอดปีออก · เลข = ตั้งปีใหม่
// **ห้ามเปลี่ยนไปใช้ `**int`** — encoding/json เจอ null แล้วเซ็ต pointer ตัวนอก
// เป็น nil ทำให้ "null" กับ "ไม่ส่ง" แยกไม่ออก (เคยเป็นบั๊กจริงที่ UpdateItem)
func UpdateContentGroup(c *gin.Context) {
	var group model.ContentGroup
	if err := database.DB.First(&group, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบหัวข้อนี้"})
		return
	}

	var body struct {
		Title     string          `json:"title" binding:"required"`
		Body      string          `json:"body"`
		Year      json.RawMessage `json:"year"`
		SortOrder int             `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	group.Title = strings.TrimSpace(body.Title)
	group.Body = body.Body
	group.SortOrder = body.SortOrder

	if len(body.Year) > 0 {
		if raw := strings.TrimSpace(string(body.Year)); raw == "null" {
			group.Year = nil
		} else {
			var y int
			if err := json.Unmarshal(body.Year, &y); err != nil || !validBuddhistYear(y) {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ปีต้องเป็น พ.ศ. เช่น 2569"})
				return
			}
			group.Year = &y
		}
	}

	// ใช้ Save เพราะต้องเขียน year = NULL ได้ด้วย
	// (Updates ด้วย struct จะข้าม zero value ทำให้ถอดปีออกไม่ได้)
	if err := database.DB.Save(&group).Error; err != nil {
		dbError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": group})
}

// DeleteContentGroup godoc
// @Summary  ลบหัวข้อ พร้อมไฟล์ที่แนบอยู่ใต้หัวข้อนั้น
// @Tags     content
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "group id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /content/groups/{id} [delete]
func DeleteContentGroup(c *gin.Context) {
	var group model.ContentGroup
	if err := database.DB.First(&group, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบหัวข้อนี้"})
		return
	}

	var urls []string
	if err := database.DB.Model(&model.ContentFile{}).
		Where("group_id = ?", group.ID).Pluck("file_url", &urls).Error; err != nil {
		dbError(c, err)
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", group.ID).Delete(&model.ContentFile{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.ContentGroup{}, group.ID).Error
	})
	if err != nil {
		dbError(c, err)
		return
	}

	removeFilesOnDisk(urls)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ---------- file (admin) ----------

// AddContentFiles godoc
// @Summary  แนบไฟล์เข้าหัวข้อ (ครั้งละหลายไฟล์ได้)
// @Tags     content
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    id     path     int    true  "group id"
// @Param    files  formData file   true  "ไฟล์ .pdf/.jpg/.jpeg/.png (ส่งซ้ำ field เดิมได้หลายไฟล์)"
// @Param    labels formData string false "ชื่อที่จะแสดงของแต่ละไฟล์ เรียงตรงลำดับกับ files"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /content/groups/{id}/files [post]
//
// ตรวจครบทุกไฟล์ก่อนเซฟสักไฟล์ (all-or-nothing) — ไฟล์เสียใบเดียวจะได้ไม่ทิ้งขยะบนดิสก์
// และ admin ไม่ต้องเดาว่าไฟล์ไหนเข้าไม่เข้า (กติกาเดียวกับอัลบั้มกิจกรรม)
func AddContentFiles(c *gin.Context) {
	maxBody := MaxPDFBytes*int64(maxContentFilesPerRequest) + maxFormOverhead
	if err := enforceUploadLimit(c, maxBody); err != nil {
		uploadLimitError(c, err)
		return
	}

	var group model.ContentGroup
	if err := database.DB.First(&group, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบหัวข้อนี้"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ไม่ได้แนบไฟล์มา"})
		return
	}
	if len(files) > maxContentFilesPerRequest {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("แนบได้ครั้งละไม่เกิน %d ไฟล์", maxContentFilesPerRequest),
		})
		return
	}
	labels := form.Value["labels"]

	exts := make([]string, len(files))
	for i, f := range files {
		ext, verr := validateUpload(f, allowedITAExt, itaMaxBytes(f.Filename))
		if verr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": fmt.Sprintf("ไฟล์ %q: %s", filepath.Base(f.Filename), verr.Error()),
			})
			return
		}
		exts[i] = ext
	}

	if err := os.MkdirAll(contentFileDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
		return
	}

	var maxOrder int
	database.DB.Model(&model.ContentFile{}).
		Where("group_id = ?", group.ID).
		Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)

	saved := make([]string, 0, len(files))
	rows := make([]model.ContentFile, 0, len(files))
	for i, f := range files {
		newFileName := fmt.Sprintf("%d-%d%s", time.Now().UnixNano(), i, exts[i])
		if err := c.SaveUploadedFile(f, filepath.Join(contentFileDir, newFileName)); err != nil {
			removeFilesOnDisk(saved)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกไฟล์ไม่สำเร็จ"})
			return
		}
		url := "/" + contentFileDir + "/" + newFileName
		saved = append(saved, url)

		// ไม่ส่ง label มา = ใช้ชื่อไฟล์เดิม ดีกว่าปล่อยว่างจนคนกดไม่ถูก
		label := strings.TrimSpace(filepath.Base(f.Filename))
		if i < len(labels) && strings.TrimSpace(labels[i]) != "" {
			label = strings.TrimSpace(labels[i])
		}

		rows = append(rows, model.ContentFile{
			GroupID:   group.ID,
			Label:     label,
			FileURL:   url,
			SortOrder: maxOrder + i + 1,
		})
	}

	if err := database.DB.Create(&rows).Error; err != nil {
		removeFilesOnDisk(saved)
		dbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": rows})
}

// UpdateContentFile godoc
// @Summary  แก้ชื่อที่แสดง/ลำดับของไฟล์ (ไม่เปลี่ยนตัวไฟล์)
// @Tags     content
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id   path int true "file id"
// @Param    body body map[string]interface{} true "label, sort_order"
// @Success  200 {object} map[string]interface{}
// @Failure  400 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /content/files/{id} [put]
func UpdateContentFile(c *gin.Context) {
	var file model.ContentFile
	if err := database.DB.First(&file, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบไฟล์นี้"})
		return
	}

	var body struct {
		Label     string `json:"label" binding:"required"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	file.Label = strings.TrimSpace(body.Label)
	file.SortOrder = body.SortOrder
	if err := database.DB.Save(&file).Error; err != nil {
		dbError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": file})
}

// DeleteContentFile godoc
// @Summary  ลบไฟล์ออกจากหัวข้อ
// @Tags     content
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "file id"
// @Success  200 {object} map[string]interface{}
// @Failure  404 {object} map[string]interface{}
// @Router   /content/files/{id} [delete]
func DeleteContentFile(c *gin.Context) {
	var file model.ContentFile
	if err := database.DB.First(&file, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบไฟล์นี้"})
		return
	}

	if err := database.DB.Delete(&model.ContentFile{}, file.ID).Error; err != nil {
		dbError(c, err)
		return
	}
	removeFilesOnDisk([]string{file.FileURL})

	c.JSON(http.StatusOK, gin.H{"success": true})
}
