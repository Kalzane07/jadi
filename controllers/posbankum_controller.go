package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================== INDEX ==================
func PosbankumIndex(c *gin.Context) {
	// ambil query search (pakai param "q")
	search := c.Query("q")

	// pagination setup
	limit := 50
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var posbankums []models.Posbankum
	db := config.DB.Model(&models.Posbankum{}).
		Preload("Kelurahan").
		Preload("Kelurahan.Kecamatan").
		Preload("Kelurahan.Kecamatan.Kabupaten")

	// filter kalau ada search
	if search != "" {
		db = db.Joins("JOIN kelurahans ON kelurahans.id = posbankums.kelurahan_id").
			Where("kelurahans.name LIKE ? OR posbankums.catatan LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// hitung total
	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error hitung total")
		return
	}

	// ambil data dengan limit + offset
	if err := db.Offset(offset).Limit(limit).Find(&posbankums).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error ambil data")
		return
	}

	// hitung total halaman
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.HTML(http.StatusOK, "posbankum_index.html", gin.H{
		"Title":      "Data Posbankum",
		"Posbankums": posbankums,
		"Search":     search,
		"Page":       page,
		"Offset":     offset,
		"TotalPages": totalPages,
	})
}

// ================== CREATE FORM ==================
func PosbankumCreate(c *gin.Context) {
	c.HTML(http.StatusOK, "posbankum_create.html", gin.H{
		"Title": "Tambah Posbankum",
	})
}

// ================== STORE ==================
func PosbankumStore(c *gin.Context) {
	kelurahanID, _ := strconv.Atoi(c.PostForm("kelurahan_id"))
	catatan := c.PostForm("catatan")

	// cek duplikasi
	var existing models.Posbankum
	if err := config.DB.Where("kelurahan_id = ?", kelurahanID).First(&existing).Error; err == nil {
		c.HTML(http.StatusOK, "posbankum_create.html", gin.H{
			"Title":          "Tambah Posbankum",
			"ErrorKelurahan": "❌ Posbankum untuk kelurahan ini sudah ada",
			"Catatan":        catatan,
		})
		return
	}

	file, err := c.FormFile("dokumen")
	if err != nil {
		c.HTML(http.StatusOK, "posbankum_create.html", gin.H{
			"Title":     "Tambah Posbankum",
			"ErrorFile": "❌ Dokumen wajib diupload",
			"Catatan":   catatan,
		})
		return
	}

	// validasi file
	if file.Size > 10*1024*1024 {
		c.HTML(http.StatusOK, "posbankum_create.html", gin.H{
			"Title":     "Tambah Posbankum",
			"ErrorFile": "❌ Ukuran file maksimal 10MB",
			"Catatan":   catatan,
		})
		return
	}
	if strings.ToLower(filepath.Ext(file.Filename)) != ".pdf" {
		c.HTML(http.StatusOK, "posbankum_create.html", gin.H{
			"Title":     "Tambah Posbankum",
			"ErrorFile": "❌ File harus berupa PDF",
			"Catatan":   catatan,
		})
		return
	}

	uploadPath := "uploads/posbankum"
	os.MkdirAll(uploadPath, os.ModePerm)

	// generate nama file unik
	ext := filepath.Ext(file.Filename)
	newName := uuid.New().String() + ext
	fullPath := filepath.Join(uploadPath, newName)

	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		c.HTML(http.StatusOK, "posbankum_create.html", gin.H{
			"Title":     "Tambah Posbankum",
			"ErrorFile": "❌ Gagal upload file",
			"Catatan":   catatan,
		})
		return
	}

	publicPath := strings.ReplaceAll(fullPath, "\\", "/")

	posbankum := models.Posbankum{
		KelurahanID: uint(kelurahanID),
		Dokumen:     publicPath,
		Catatan:     catatan,
	}

	config.DB.Create(&posbankum)
	c.Redirect(http.StatusFound, "/jadi/admin/posbankum")
}

// ================== EDIT FORM ==================
func PosbankumEdit(c *gin.Context) {
	id := c.Param("id")
	var posbankum models.Posbankum
	if err := config.DB.
		Preload("Kelurahan").
		Preload("Kelurahan.Kecamatan").
		Preload("Kelurahan.Kecamatan.Kabupaten").
		First(&posbankum, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "Data tidak ditemukan")
		} else {
			c.String(http.StatusInternalServerError, "Error DB")
		}
		return
	}

	c.HTML(http.StatusOK, "posbankum_edit.html", gin.H{
		"Title":     "Edit Posbankum",
		"Posbankum": posbankum,
	})
}

// ================== UPDATE ==================
func PosbankumUpdate(c *gin.Context) {
	id := c.Param("id")
	var posbankum models.Posbankum
	if err := config.DB.First(&posbankum, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	kelurahanID, _ := strconv.Atoi(c.PostForm("kelurahan_id"))

	// cek duplikasi selain dirinya sendiri
	var count int64
	config.DB.Model(&models.Posbankum{}).
		Where("kelurahan_id = ? AND id <> ?", kelurahanID, posbankum.ID).
		Count(&count)
	if count > 0 {
		c.HTML(http.StatusOK, "posbankum_edit.html", gin.H{
			"Title":          "Edit Posbankum",
			"Posbankum":      posbankum,
			"ErrorKelurahan": "❌ Posbankum untuk kelurahan ini sudah ada",
		})
		return
	}

	posbankum.KelurahanID = uint(kelurahanID)
	posbankum.Catatan = c.PostForm("catatan")

	// cek file baru
	file, err := c.FormFile("dokumen")
	if err == nil {
		if file.Size > 10*1024*1024 {
			c.HTML(http.StatusOK, "posbankum_edit.html", gin.H{
				"Title":     "Edit Posbankum",
				"Posbankum": posbankum,
				"ErrorFile": "❌ Ukuran file maksimal 10MB",
			})
			return
		}
		if strings.ToLower(filepath.Ext(file.Filename)) != ".pdf" {
			c.HTML(http.StatusOK, "posbankum_edit.html", gin.H{
				"Title":     "Edit Posbankum",
				"Posbankum": posbankum,
				"ErrorFile": "❌ File harus berupa PDF",
			})
			return
		}

		uploadPath := "uploads/posbankum"
		os.MkdirAll(uploadPath, os.ModePerm)

		// generate nama file unik
		ext := filepath.Ext(file.Filename)
		newName := uuid.New().String() + ext
		newPath := filepath.Join(uploadPath, newName)

		if err := c.SaveUploadedFile(file, newPath); err != nil {
			c.HTML(http.StatusOK, "posbankum_edit.html", gin.H{
				"Title":     "Edit Posbankum",
				"Posbankum": posbankum,
				"ErrorFile": "❌ Gagal upload file",
			})
			return
		}

		// hapus file lama kalau ada
		if posbankum.Dokumen != "" {
			_ = os.Remove(posbankum.Dokumen)
		}

		posbankum.Dokumen = strings.ReplaceAll(newPath, "\\", "/")
	}

	config.DB.Save(&posbankum)
	c.Redirect(http.StatusFound, "/jadi/admin/posbankum")
}

// ================== DELETE ==================
func PosbankumDelete(c *gin.Context) {
	id := c.Param("id")
	var posbankum models.Posbankum

	if err := config.DB.First(&posbankum, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	// hapus file dari storage kalau ada
	if posbankum.Dokumen != "" {
		_ = os.Remove(posbankum.Dokumen)
	}

	// hapus record dari DB
	config.DB.Delete(&posbankum)

	c.Redirect(http.StatusFound, "/jadi/admin/posbankum")
}

// ================== API: Autocomplete Kelurahan ==================
func KelurahanSearch(c *gin.Context) {
	term := c.Query("term")
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "term required"})
		return
	}

	var kelurahans []models.Kelurahan
	config.DB.
		Preload("Kecamatan").
		Preload("Kecamatan.Kabupaten").
		Where("name LIKE ?", "%"+strings.TrimSpace(term)+"%").
		Limit(20).
		Find(&kelurahans)

	results := []gin.H{}
	for _, k := range kelurahans {
		results = append(results, gin.H{
			"id":        k.ID,
			"name":      k.Name,
			"kecamatan": k.Kecamatan.Name,
			"kabupaten": k.Kecamatan.Kabupaten.Name,
		})
	}

	c.JSON(http.StatusOK, results)
}
