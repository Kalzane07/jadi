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
func PJAIndex(c *gin.Context) {
	search := c.Query("q")

	limit := 50
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var pjas []models.Pja
	db := config.DB.Model(&models.Pja{}).
		Preload("Kelurahan").
		Preload("Kelurahan.Kecamatan").
		Preload("Kelurahan.Kecamatan.Kabupaten")

	if search != "" {
		db = db.Joins("JOIN kelurahans ON kelurahans.id = pjas.kelurahan_id").
			Where("kelurahans.name LIKE ? OR pjas.catatan LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error hitung total")
		return
	}

	if err := db.Offset(offset).Limit(limit).Find(&pjas).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error ambil data")
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.HTML(http.StatusOK, "pja_index.html", gin.H{
		"Title":      "Data PJA",
		"Pjas":       pjas,
		"Search":     search,
		"Page":       page,
		"Offset":     offset,
		"TotalPages": totalPages,
	})
}

// ================== CREATE FORM ==================
func PJACreate(c *gin.Context) {
	c.HTML(http.StatusOK, "pja_create.html", gin.H{
		"Title": "Tambah PJA",
	})
}

// ================== STORE ==================
func PJAStore(c *gin.Context) {
	kelurahanID, _ := strconv.Atoi(c.PostForm("kelurahan_id"))
	catatan := utils.SanitizeInput(c.PostForm("catatan"))
	// Cek duplikasi kelurahan
	var existing models.Pja
	if err := config.DB.Where("kelurahan_id = ?", kelurahanID).First(&existing).Error; err == nil {
		c.HTML(http.StatusOK, "pja_create.html", gin.H{
			"Title":          "Tambah PJA",
			"ErrorKelurahan": "❌ PJA untuk kelurahan ini sudah ada",
			"Catatan":        catatan,
		})
		return
	}

	// Validasi file upload
	file, err := c.FormFile("dokumen")
	if err != nil {
		c.HTML(http.StatusOK, "pja_create.html", gin.H{
			"Title":     "Tambah PJA",
			"ErrorFile": "❌ Dokumen wajib diupload",
			"Catatan":   catatan,
		})
		return
	}
	if !utils.ValidatePDFUpload(c, file) {
		c.HTML(http.StatusOK, "pja_create.html", gin.H{
			"Title":     "Tambah PJA",
			"ErrorFile": "❌ File tidak valid. Pastikan file adalah PDF dan ukurannya di bawah 10MB.",
			"Catatan":   catatan,
		})
		return
	}

	uploadPath := "uploads/pja"
	os.MkdirAll(uploadPath, os.ModePerm)

	// Generate nama file unik menggunakan uuid
	ext := filepath.Ext(file.Filename)
	newName := uuid.New().String() + ext
	fullPath := filepath.Join(uploadPath, newName)

	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		c.HTML(http.StatusOK, "pja_create.html", gin.H{
			"Title":     "Tambah PJA",
			"ErrorFile": "❌ Gagal upload file",
			"Catatan":   catatan,
		})
		return
	}

	// Buat record baru
	publicPath := strings.ReplaceAll(fullPath, "\\", "/")
	pja := models.Pja{
		KelurahanID: uint(kelurahanID),
		Dokumen:     publicPath,
		Catatan:     catatan,
	}

	config.DB.Create(&pja)
	c.Redirect(http.StatusFound, "/admin/pja")
}

func PJAView(c *gin.Context) {
	id := c.Param("id")
	var pja models.Pja
	if err := config.DB.
		Preload("Kelurahan").
		Preload("Kelurahan.Kecamatan").
		Preload("Kelurahan.Kecamatan.Kabupaten").
		First(&pja, id).Error; err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}
	c.Header("Content-Disposition", "inline; filename="+filepath.Base(pja.Dokumen))
	c.Header("Content-Type", "application/pdf")
	c.File(pja.Dokumen)
}

// ================== EDIT FORM ==================
func PJAEdit(c *gin.Context) {
	id := c.Param("id")
	var pja models.Pja
	if err := config.DB.
		Preload("Kelurahan").
		Preload("Kelurahan.Kecamatan").
		Preload("Kelurahan.Kecamatan.Kabupaten").
		First(&pja, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "Data tidak ditemukan")
		} else {
			c.String(http.StatusInternalServerError, "Error DB")
		}
		return
	}

	c.HTML(http.StatusOK, "pja_edit.html", gin.H{
		"Title": "Edit PJA",
		"PJA":   pja,
	})
}

// ================== UPDATE ==================
func PJAUpdate(c *gin.Context) {
	id := c.Param("id")
	var pja models.Pja
	if err := config.DB.First(&pja, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	kelurahanID, _ := strconv.Atoi(c.PostForm("kelurahan_id"))

	var count int64
	config.DB.Model(&models.Pja{}).
		Where("kelurahan_id = ? AND id <> ?", kelurahanID, pja.ID).
		Count(&count)
	if count > 0 {
		c.HTML(http.StatusOK, "pja_edit.html", gin.H{
			"Title":          "Edit PJA",
			"PJA":            pja,
			"ErrorKelurahan": "❌ PJA untuk kelurahan ini sudah ada",
		})
		return
	}

	// Update field
	pja.KelurahanID = uint(kelurahanID)
	pja.Catatan = utils.SanitizeInput(c.PostForm("catatan"))

	file, err := c.FormFile("dokumen")
	// Jika ada file baru yang diupload
	if err == nil {
		if !utils.ValidatePDFUpload(c, file) {
			c.HTML(http.StatusOK, "pja_edit.html", gin.H{
				"Title":     "Edit PJA",
				"PJA":       pja,
				"ErrorFile": "❌ File tidak valid. Pastikan file adalah PDF dan ukurannya di bawah 10MB.",
			})
			return
		}

		uploadPath := "uploads/pja"
		os.MkdirAll(uploadPath, os.ModePerm)

		// Generate nama file unik menggunakan uuid
		ext := filepath.Ext(file.Filename)
		newName := uuid.New().String() + ext
		newPath := filepath.Join(uploadPath, newName)

		if err := c.SaveUploadedFile(file, newPath); err != nil {
			c.HTML(http.StatusOK, "pja_edit.html", gin.H{
				"Title":     "Edit PJA",
				"PJA":       pja,
				"ErrorFile": "❌ Gagal upload file",
			})
			return
		}

		// Hapus file lama jika ada
		if pja.Dokumen != "" {
			_ = os.Remove(pja.Dokumen)
		}

		pja.Dokumen = strings.ReplaceAll(newPath, "\\", "/")
	}

	config.DB.Save(&pja)
	c.Redirect(http.StatusFound, "/admin/pja")
}

// ================== DELETE ==================
func PJADelete(c *gin.Context) {
	id := c.Param("id")
	var pja models.Pja

	if err := config.DB.First(&pja, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	// hapus file PDF kalau ada
	if pja.Dokumen != "" {
		_ = os.Remove(pja.Dokumen)
	}

	// hapus record
	config.DB.Delete(&pja)

	c.Redirect(http.StatusFound, "/admin/pja")
}

// ================== API: Autocomplete Kelurahan ==================
func PJAKelurahanSearch(c *gin.Context) {
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
		Or("kode LIKE ?", "%"+strings.TrimSpace(term)+"%").
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
