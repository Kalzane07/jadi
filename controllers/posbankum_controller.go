package controllers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================== INDEX ==================
func PosbankumIndex(c *gin.Context) {
	search := c.Query("q")
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

	if search != "" {
		db = db.Joins("JOIN kelurahans ON kelurahans.id = posbankums.kelurahan_id").
			Where("kelurahans.name LIKE ? OR posbankums.catatan LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error menghitung total data")
		return
	}

	if err := db.Offset(offset).Limit(limit).Find(&posbankums).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error mengambil data")
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Mengambil username dari session untuk ditampilkan di navbar
	session := sessions.Default(c)
	user := session.Get("user")

	c.HTML(http.StatusOK, "posbankum_index.html", gin.H{
		"Title":      "Data Posbankum",
		"Posbankums": posbankums,
		"Search":     search,
		"Page":       page,
		"Offset":     offset,
		"TotalPages": totalPages,
		"user":       user,
	})
}

// ================== CREATE FORM ==================
func PosbankumCreate(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	c.HTML(http.StatusOK, "posbankum_create.html", gin.H{
		"Title": "Tambah Posbankum",
		"user":  user,
	})
}

// ================== STORE ==================
func PosbankumStore(c *gin.Context) {
	kelurahanID, _ := strconv.Atoi(c.PostForm("kelurahan_id"))
	catatan := c.PostForm("catatan")

	var existing models.Posbankum
	if err := config.DB.Where("kelurahan_id = ?", kelurahanID).First(&existing).Error; err == nil {
		c.HTML(http.StatusBadRequest, "posbankum_create.html", gin.H{
			"Title":          "Tambah Posbankum",
			"ErrorKelurahan": "Posbankum untuk kelurahan ini sudah ada.",
			"Catatan":        catatan,
		})
		return
	}

	file, err := c.FormFile("dokumen")
	if err != nil {
		c.HTML(http.StatusBadRequest, "posbankum_create.html", gin.H{
			"Title":     "Tambah Posbankum",
			"ErrorFile": "Dokumen wajib diunggah.",
			"Catatan":   catatan,
		})
		return
	}

	if file.Size > 10*1024*1024 { // 10MB
		c.HTML(http.StatusBadRequest, "posbankum_create.html", gin.H{
			"Title":     "Tambah Posbankum",
			"ErrorFile": "Ukuran file maksimal 10MB.",
			"Catatan":   catatan,
		})
		return
	}
	if strings.ToLower(filepath.Ext(file.Filename)) != ".pdf" {
		c.HTML(http.StatusBadRequest, "posbankum_create.html", gin.H{
			"Title":     "Tambah Posbankum",
			"ErrorFile": "File harus berformat PDF.",
			"Catatan":   catatan,
		})
		return
	}

	uploadPath := "uploads/posbankum"
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		c.String(http.StatusInternalServerError, "Gagal membuat direktori upload")
		return
	}

	ext := filepath.Ext(file.Filename)
	newName := uuid.New().String() + ext
	fullPath := filepath.Join(uploadPath, newName)

	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		c.String(http.StatusInternalServerError, "Gagal menyimpan file")
		return
	}

	publicPath := strings.ReplaceAll(fullPath, "\\", "/")
	posbankum := models.Posbankum{
		KelurahanID: uint(kelurahanID),
		Dokumen:     publicPath,
		Catatan:     catatan,
	}

	if err := config.DB.Create(&posbankum).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal menyimpan data ke database")
		return
	}
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
			c.String(http.StatusInternalServerError, "Error database")
		}
		return
	}

	session := sessions.Default(c)
	user := session.Get("user")

	c.HTML(http.StatusOK, "posbankum_edit.html", gin.H{
		"Title":     "Edit Posbankum",
		"Posbankum": posbankum,
		"user":      user,
	})
}

// ================== UPDATE ==================
func PosbankumUpdate(c *gin.Context) {
	id := c.Param("id")
	var posbankum models.Posbankum

	// OPTIMISASI: Mengambil data lengkap dengan preload cukup satu kali di awal.
	if err := config.DB.
		Preload("Kelurahan").
		Preload("Kelurahan.Kecamatan").
		Preload("Kelurahan.Kecamatan.Kabupaten").
		First(&posbankum, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "Data tidak ditemukan")
		} else {
			c.String(http.StatusInternalServerError, "Error database")
		}
		return
	}

	kelurahanID, _ := strconv.Atoi(c.PostForm("kelurahan_id"))
	posbankum.KelurahanID = uint(kelurahanID)
	posbankum.Catatan = c.PostForm("catatan")

	// Cek duplikasi kelurahan selain data itu sendiri
	var count int64
	config.DB.Model(&models.Posbankum{}).
		Where("kelurahan_id = ? AND id <> ?", kelurahanID, posbankum.ID).
		Count(&count)

	if count > 0 {
		session := sessions.Default(c)
		user := session.Get("user")
		c.HTML(http.StatusBadRequest, "posbankum_edit.html", gin.H{
			"Title":          "Edit Posbankum",
			"Posbankum":      posbankum,
			"ErrorKelurahan": "Posbankum untuk kelurahan ini sudah ada.",
			"user":           user,
		})
		return
	}

	file, err := c.FormFile("dokumen")
	if err == nil { // Ada file baru yang diupload
		if file.Size > 10*1024*1024 {
			session := sessions.Default(c)
			user := session.Get("user")
			c.HTML(http.StatusBadRequest, "posbankum_edit.html", gin.H{
				"Title":     "Edit Posbankum",
				"Posbankum": posbankum,
				"ErrorFile": "Ukuran file maksimal 10MB.",
				"user":      user,
			})
			return
		}
		if strings.ToLower(filepath.Ext(file.Filename)) != ".pdf" {
			session := sessions.Default(c)
			user := session.Get("user")
			c.HTML(http.StatusBadRequest, "posbankum_edit.html", gin.H{
				"Title":     "Edit Posbankum",
				"Posbankum": posbankum,
				"ErrorFile": "File harus berformat PDF.",
				"user":      user,
			})
			return
		}

		uploadPath := "uploads/posbankum"
		os.MkdirAll(uploadPath, os.ModePerm)

		ext := filepath.Ext(file.Filename)
		newName := uuid.New().String() + ext
		newPath := filepath.Join(uploadPath, newName)

		if err := c.SaveUploadedFile(file, newPath); err != nil {
			c.String(http.StatusInternalServerError, "Gagal menyimpan file baru")
			return
		}

		if posbankum.Dokumen != "" {
			if err := os.Remove(posbankum.Dokumen); err != nil {
				log.Printf("Gagal menghapus file lama: %s, error: %v", posbankum.Dokumen, err)
			}
		}

		posbankum.Dokumen = strings.ReplaceAll(newPath, "\\", "/")
	}

	if err := config.DB.Save(&posbankum).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal mengupdate data di database")
		return
	}
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

	if posbankum.Dokumen != "" {
		if err := os.Remove(posbankum.Dokumen); err != nil {
			log.Printf("Gagal menghapus file: %s, error: %v", posbankum.Dokumen, err)
		}
	}

	if err := config.DB.Delete(&posbankum).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal menghapus data dari database")
		return
	}

	c.Redirect(http.StatusFound, "/jadi/admin/posbankum")
}

// ================== API: Autocomplete Kelurahan ==================
func KelurahanSearch(c *gin.Context) {
	term := c.Query("term")
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "term is required"})
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
