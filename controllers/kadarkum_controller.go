package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-gonic/gin"
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
	catatan := c.PostForm("catatan")

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
	if file.Size > 10*1024*1024 {
		c.HTML(http.StatusOK, "kadarkum_create.html", gin.H{
			"Title":     "Tambah Kadarkum",
			"ErrorFile": "❌ Ukuran file maksimal 10MB",
			"Catatan":   catatan,
		})
		return
	}
	if strings.ToLower(filepath.Ext(file.Filename)) != ".pdf" {
		c.HTML(http.StatusOK, "kadarkum_create.html", gin.H{
			"Title":     "Tambah Kadarkum",
			"ErrorFile": "❌ File harus berupa PDF",
			"Catatan":   catatan,
		})
		return
	}

	uploadPath := "uploads/kadarkum"
	os.MkdirAll(uploadPath, os.ModePerm)

	// generate nama file unik (timestamp + random)
	ext := filepath.Ext(file.Filename)
	newName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Intn(1000), ext)
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
	kadarkum.Catatan = c.PostForm("catatan")

	file, err := c.FormFile("dokumen")
	if err == nil {
		if file.Size > 10*1024*1024 {
			c.HTML(http.StatusOK, "kadarkum_edit.html", gin.H{
				"Title":     "Edit Kadarkum",
				"Kadarkum":  kadarkum,
				"ErrorFile": "❌ Ukuran file maksimal 10MB",
			})
			return
		}
		if strings.ToLower(filepath.Ext(file.Filename)) != ".pdf" {
			c.HTML(http.StatusOK, "kadarkum_edit.html", gin.H{
				"Title":     "Edit Kadarkum",
				"Kadarkum":  kadarkum,
				"ErrorFile": "❌ File harus berupa PDF",
			})
			return
		}

		uploadPath := "uploads/kadarkum"
		os.MkdirAll(uploadPath, os.ModePerm)

		ext := filepath.Ext(file.Filename)
		newName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Intn(1000), ext)
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

	// hapus file kalau ada
	if kadarkum.Dokumen != "" {
		_ = os.Remove(kadarkum.Dokumen)
	}

	config.DB.Delete(&kadarkum)
	c.Redirect(http.StatusFound, "/jadi/admin/kadarkum")
}
