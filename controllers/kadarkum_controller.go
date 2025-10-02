package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go-admin/config"
	"go-admin/models"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================== INDEX ==================
func KadarkumIndex(c *gin.Context) {
	search := c.Query("q")

	// pagination
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
		c.String(http.StatusInternalServerError, "Error hitung total")
		return
	}

	if err := db.Offset(offset).Limit(limit).Find(&kadarkums).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error ambil data")
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.HTML(http.StatusOK, "kadarkum_index.html", gin.H{
		"Title":      "Data Kadarkum",
		"Kadarkums":  kadarkums,
		"Search":     search,
		"Page":       page,
		"Offset":     offset,
		"TotalPages": totalPages,
	})
}

// ================== CREATE FORM ==================
func KadarkumCreate(c *gin.Context) {
	c.HTML(http.StatusOK, "kadarkum_create.html", gin.H{
		"Title": "Tambah Kadarkum",
	})
}

// ================== STORE ==================
func KadarkumStore(c *gin.Context) {
	kelurahanID, _ := strconv.Atoi(c.PostForm("kelurahan_id"))

	catatan := utils.SanitizeInput(c.PostForm("catatan"))
	// cek duplikasi
	var existing models.Kadarkum
	if err := config.DB.Where("kelurahan_id = ?", kelurahanID).First(&existing).Error; err == nil {
		c.HTML(http.StatusOK, "kadarkum_create.html", gin.H{
			"Title":          "Tambah Kadarkum",
			"ErrorKelurahan": "❌ Kadarkum untuk kelurahan ini sudah ada",
			"Catatan":        catatan,
		})
		return
	}

	file, err := c.FormFile("dokumen")
	if err != nil {
		c.HTML(http.StatusOK, "kadarkum_create.html", gin.H{
			"Title":     "Tambah Kadarkum",
			"ErrorFile": "❌ Dokumen wajib diupload",
			"Catatan":   catatan,
		})
		return
	}

	// validasi file
	if !utils.ValidatePDFUpload(c, file) {
		c.HTML(http.StatusOK, "kadarkum_create.html", gin.H{
			"Title":     "Tambah Kadarkum",
			"ErrorFile": "❌ File tidak valid. Pastikan file adalah PDF dan ukurannya di bawah 10MB.",
			"Catatan":   catatan,
		})
		return
	}

	uploadPath := "uploads/kadarkum"
	os.MkdirAll(uploadPath, os.ModePerm)

	// generate nama file unik menggunakan uuid
	ext := filepath.Ext(file.Filename)
	newName := uuid.New().String() + ext
	fullPath := filepath.Join(uploadPath, newName)

	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		c.HTML(http.StatusOK, "kadarkum_create.html", gin.H{
			"Title":     "Tambah Kadarkum",
			"ErrorFile": "❌ Gagal upload file",
			"Catatan":   catatan,
		})
		return
	}

	publicPath := strings.ReplaceAll(fullPath, "\\", "/")

	kadarkum := models.Kadarkum{
		KelurahanID: uint(kelurahanID),
		Dokumen:     publicPath,
		Catatan:     catatan,
	}

	config.DB.Create(&kadarkum)
	c.Redirect(http.StatusFound, "/admin/kadarkum")
}

func KadarkumView(c *gin.Context) {
	id := c.Param("id")
	var kadarkum models.Kadarkum
	if err := config.DB.First(&kadarkum, id).Error; err != nil {
		c.String(http.StatusNotFound, "Dokumen tidak ditemukan: "+err.Error())
		return
	}
	c.Header("Content-Disposition", "inline; filename="+filepath.Base(kadarkum.Dokumen))
	c.Header("Content-Type", "application/pdf")
	c.File(kadarkum.Dokumen)
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
			c.String(http.StatusInternalServerError, "Error DB")
		}
		return
	}

	c.HTML(http.StatusOK, "kadarkum_edit.html", gin.H{
		"Title":    "Edit Kadarkum",
		"Kadarkum": kadarkum,
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

	// cek duplikasi selain dirinya sendiri
	var count int64
	config.DB.Model(&models.Kadarkum{}).
		Where("kelurahan_id = ? AND id <> ?", kelurahanID, kadarkum.ID).
		Count(&count)
	if count > 0 {
		c.HTML(http.StatusOK, "kadarkum_edit.html", gin.H{
			"Title":          "Edit Kadarkum",
			"Kadarkum":       kadarkum,
			"ErrorKelurahan": "❌ Kadarkum untuk kelurahan ini sudah ada",
		})
		return
	}

	kadarkum.KelurahanID = uint(kelurahanID)

	kadarkum.Catatan = utils.SanitizeInput(c.PostForm("catatan"))

	file, err := c.FormFile("dokumen")
	if err == nil {
		if !utils.ValidatePDFUpload(c, file) {
			c.HTML(http.StatusOK, "kadarkum_edit.html", gin.H{
				"Title":     "Edit Kadarkum",
				"Kadarkum":  kadarkum,
				"ErrorFile": "❌ File tidak valid. Pastikan file adalah PDF dan ukurannya di bawah 10MB.",
			})
			return
		}

		uploadPath := "uploads/kadarkum"
		os.MkdirAll(uploadPath, os.ModePerm)

		// generate nama file unik menggunakan uuid
		ext := filepath.Ext(file.Filename)
		newName := uuid.New().String() + ext
		newPath := filepath.Join(uploadPath, newName)

		if err := c.SaveUploadedFile(file, newPath); err != nil {
			c.HTML(http.StatusOK, "kadarkum_edit.html", gin.H{
				"Title":     "Edit Kadarkum",
				"Kadarkum":  kadarkum,
				"ErrorFile": "❌ Gagal upload file",
			})
			return
		}

		// hapus file lama kalau ada
		if kadarkum.Dokumen != "" {
			_ = os.Remove(kadarkum.Dokumen)
		}

		kadarkum.Dokumen = strings.ReplaceAll(newPath, "\\", "/")
	}

	config.DB.Save(&kadarkum)
	c.Redirect(http.StatusFound, "/admin/kadarkum")
}

// ================== DELETE ==================
func KadarkumDelete(c *gin.Context) {
	id := c.Param("id")
	var kadarkum models.Kadarkum

	if err := config.DB.First(&kadarkum, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	// hapus file kalau ada
	if kadarkum.Dokumen != "" {
		_ = os.Remove(kadarkum.Dokumen)
	}

	config.DB.Delete(&kadarkum)
	c.Redirect(http.StatusFound, "/admin/kadarkum")
}
