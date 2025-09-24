package controllers

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ================== INDEX ==================
func KadarkumIndex(c *gin.Context) {
	search := c.Query("q")
	limit := 50
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var kadarkums []models.Kadarkum
	db := config.DB.Model(&models.Kadarkum{}).
		Preload("Kelurahan").
		Preload("Kelurahan.Kecamatan").
		Preload("Kelurahan.Kecamatan.Kabupaten")

	if search != "" {
		db = db.Joins("JOIN kelurahans ON kelurahans.id = kadarkums.kelurahan_id").
			Where("kelurahans.name LIKE ? OR kadarkums.catatan LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error menghitung total data")
		return
	}

	if err := db.Offset(offset).Limit(limit).Find(&kadarkums).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error mengambil data")
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	session := sessions.Default(c)
	user := session.Get("user")

	c.HTML(http.StatusOK, "kadarkum_index.html", gin.H{
		"Title":      "Data Kadarkum",
		"Kadarkums":  kadarkums,
		"Search":     search,
		"Page":       page,
		"Offset":     offset,
		"TotalPages": totalPages,
		"user":       user,
	})
}

// ================== CREATE FORM ==================
func KadarkumCreate(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")

	c.HTML(http.StatusOK, "kadarkum_create.html", gin.H{
		"Title": "Tambah Kadarkum",
		"user":  user,
	})
}

// ================== STORE ==================
func KadarkumStore(c *gin.Context) {
	kelurahanID, _ := strconv.Atoi(c.PostForm("kelurahan_id"))
	catatan := c.PostForm("catatan")

	// cek duplikasi
	var existing models.Kadarkum
	if err := config.DB.Where("kelurahan_id = ?", kelurahanID).First(&existing).Error; err == nil {
		c.HTML(http.StatusBadRequest, "kadarkum_create.html", gin.H{
			"Title":          "Tambah Kadarkum",
			"ErrorKelurahan": "❌ Kadarkum untuk kelurahan ini sudah ada",
			"Catatan":        catatan,
		})
		return
	}

	file, err := c.FormFile("dokumen")
	if err != nil {
		c.HTML(http.StatusBadRequest, "kadarkum_create.html", gin.H{
			"Title":     "Tambah Kadarkum",
			"ErrorFile": "❌ Dokumen wajib diupload",
			"Catatan":   catatan,
		})
		return
	}

	if file.Size > 10*1024*1024 {
		c.HTML(http.StatusBadRequest, "kadarkum_create.html", gin.H{
			"Title":     "Tambah Kadarkum",
			"ErrorFile": "❌ Ukuran file maksimal 10MB",
			"Catatan":   catatan,
		})
		return
	}

	// ✅ cek ekstensi
	if filepath.Ext(file.Filename) != ".pdf" {
		c.HTML(http.StatusBadRequest, "kadarkum_create.html", gin.H{
			"Title":     "Tambah Kadarkum",
			"ErrorFile": "❌ Ekstensi file harus .pdf",
			"Catatan":   catatan,
		})
		return
	}

	src, _ := file.Open()
	defer src.Close()

	// ✅ cek magic bytes (%PDF-)
	buf := make([]byte, 5)
	if _, err := src.Read(buf); err != nil || string(buf) != "%PDF-" {
		c.HTML(http.StatusBadRequest, "kadarkum_create.html", gin.H{
			"Title":     "Tambah Kadarkum",
			"ErrorFile": "❌ File bukan PDF valid",
			"Catatan":   catatan,
		})
		return
	}
	src.Seek(0, io.SeekStart)

	fileBytes := make([]byte, file.Size)
	src.Read(fileBytes)

	kadarkum := models.Kadarkum{
		KelurahanID: uint(kelurahanID),
		Dokumen:     fileBytes,
		ContentType: "application/pdf", // force
		Catatan:     catatan,
	}

	if err := config.DB.Create(&kadarkum).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal simpan data")
		return
	}

	c.Redirect(http.StatusFound, "/jadi/admin/kadarkum")
}

// ================== EDIT FORM ==================
func KadarkumEdit(c *gin.Context) {
	id := c.Param("id")
	var kadarkum models.Kadarkum
	if err := config.DB.
		Preload("Kelurahan").
		Preload("Kelurahan.Kecamatan").
		Preload("Kelurahan.Kecamatan.Kabupaten").
		First(&kadarkum, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "Data tidak ditemukan")
		} else {
			c.String(http.StatusInternalServerError, "Error database")
		}
		return
	}

	session := sessions.Default(c)
	user := session.Get("user")

	c.HTML(http.StatusOK, "kadarkum_edit.html", gin.H{
		"Title":    "Edit Kadarkum",
		"Kadarkum": kadarkum,
		"user":     user,
	})
}

// ================== UPDATE ==================
func KadarkumUpdate(c *gin.Context) {
	id := c.Param("id")
	var kadarkum models.Kadarkum
	if err := config.DB.First(&kadarkum, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	kelurahanID, _ := strconv.Atoi(c.PostForm("kelurahan_id"))
	catatan := c.PostForm("catatan")

	// cek duplikasi
	var count int64
	config.DB.Model(&models.Kadarkum{}).
		Where("kelurahan_id = ? AND id <> ?", kelurahanID, kadarkum.ID).
		Count(&count)
	if count > 0 {
		c.HTML(http.StatusBadRequest, "kadarkum_edit.html", gin.H{
			"Title":          "Edit Kadarkum",
			"Kadarkum":       kadarkum,
			"ErrorKelurahan": "❌ Kadarkum untuk kelurahan ini sudah ada",
		})
		return
	}

	kadarkum.KelurahanID = uint(kelurahanID)
	kadarkum.Catatan = catatan

	file, err := c.FormFile("dokumen")
	if err == nil {
		if file.Size > 10*1024*1024 {
			c.HTML(http.StatusBadRequest, "kadarkum_edit.html", gin.H{
				"Title":     "Edit Kadarkum",
				"Kadarkum":  kadarkum,
				"ErrorFile": "❌ Ukuran file maksimal 10MB",
			})
			return
		}

		if filepath.Ext(file.Filename) != ".pdf" {
			c.HTML(http.StatusBadRequest, "kadarkum_edit.html", gin.H{
				"Title":     "Edit Kadarkum",
				"Kadarkum":  kadarkum,
				"ErrorFile": "❌ Ekstensi file harus .pdf",
			})
			return
		}

		src, _ := file.Open()
		defer src.Close()

		buf := make([]byte, 5)
		if _, err := src.Read(buf); err != nil || string(buf) != "%PDF-" {
			c.HTML(http.StatusBadRequest, "kadarkum_edit.html", gin.H{
				"Title":     "Edit Kadarkum",
				"Kadarkum":  kadarkum,
				"ErrorFile": "❌ File bukan PDF valid",
			})
			return
		}
		src.Seek(0, io.SeekStart)

		fileBytes := make([]byte, file.Size)
		src.Read(fileBytes)

		kadarkum.Dokumen = fileBytes
		kadarkum.ContentType = "application/pdf"
	}

	if err := config.DB.Save(&kadarkum).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal update data")
		return
	}

	c.Redirect(http.StatusFound, "/jadi/admin/kadarkum")
}

// ================== DELETE ==================
func KadarkumDelete(c *gin.Context) {
	id := c.Param("id")
	var kadarkum models.Kadarkum

	if err := config.DB.First(&kadarkum, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	if err := config.DB.Delete(&kadarkum).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal hapus data")
		return
	}

	c.Redirect(http.StatusFound, "/jadi/admin/kadarkum")
}
