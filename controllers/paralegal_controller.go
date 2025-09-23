package controllers

import (
	"net/http"
	"strconv"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ================== INDEX ==================
func ParalegalIndex(c *gin.Context) {
	search := c.Query("q")

	limit := 50
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var paralegals []models.Paralegal
	db := config.DB.Model(&models.Paralegal{}).
		Preload("Posbankum").
		Preload("Posbankum.Kelurahan").
		Preload("Posbankum.Kelurahan.Kecamatan").
		Preload("Posbankum.Kelurahan.Kecamatan.Kabupaten")

	if search != "" {
		db = db.Joins("JOIN posbankums ON posbankums.id = paralegals.posbankum_id").
			Joins("JOIN kelurahans ON kelurahans.id = posbankums.kelurahan_id").
			Where("paralegals.nama LIKE ? OR kelurahans.name LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error hitung total")
		return
	}

	if err := db.Offset(offset).Limit(limit).Find(&paralegals).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error ambil data")
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.HTML(http.StatusOK, "paralegal_index.html", gin.H{
		"Title":      "Data Paralegal",
		"Paralegals": paralegals,
		"Search":     search,
		"Page":       page,
		"Offset":     offset,
		"TotalPages": totalPages,
	})
}

// ================== CREATE FORM ==================
func ParalegalCreate(c *gin.Context) {
	var posbankums []models.Posbankum
	config.DB.Preload("Kelurahan").Find(&posbankums)

	c.HTML(http.StatusOK, "paralegal_create.html", gin.H{
		"Title":      "Tambah Paralegal",
		"Posbankums": posbankums,
	})
}

// ================== STORE ==================
func ParalegalStore(c *gin.Context) {
	posbankumID, _ := strconv.Atoi(c.PostForm("posbankum_id"))
	nama := c.PostForm("nama")

	file, err := c.FormFile("dokumen")
	if err != nil {
		c.HTML(http.StatusBadRequest, "paralegal_create.html", gin.H{
			"Title":     "Tambah Paralegal",
			"ErrorFile": "❌ Dokumen wajib diupload",
			"Nama":      nama,
		})
		return
	}

	if file.Size > 10*1024*1024 {
		c.HTML(http.StatusBadRequest, "paralegal_create.html", gin.H{
			"Title":     "Tambah Paralegal",
			"ErrorFile": "❌ Ukuran file maksimal 10MB",
			"Nama":      nama,
		})
		return
	}

	if file.Header.Get("Content-Type") != "application/pdf" {
		c.HTML(http.StatusBadRequest, "paralegal_create.html", gin.H{
			"Title":     "Tambah Paralegal",
			"ErrorFile": "❌ File harus berupa PDF",
			"Nama":      nama,
		})
		return
	}

	f, _ := file.Open()
	defer f.Close()
	fileBytes := make([]byte, file.Size)
	f.Read(fileBytes)

	paralegal := models.Paralegal{
		PosbankumID: uint(posbankumID),
		Nama:        nama,
		Dokumen:     fileBytes,
		ContentType: file.Header.Get("Content-Type"),
	}

	if err := config.DB.Create(&paralegal).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal simpan data")
		return
	}

	c.Redirect(http.StatusFound, "/jadi/admin/paralegal")
}

// ================== EDIT FORM ==================
func ParalegalEdit(c *gin.Context) {
	id := c.Param("id")
	var paralegal models.Paralegal
	if err := config.DB.
		Preload("Posbankum").
		Preload("Posbankum.Kelurahan").
		Preload("Posbankum.Kelurahan.Kecamatan").
		Preload("Posbankum.Kelurahan.Kecamatan.Kabupaten").
		First(&paralegal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "Data tidak ditemukan")
		} else {
			c.String(http.StatusInternalServerError, "Error DB")
		}
		return
	}

	var posbankums []models.Posbankum
	config.DB.Preload("Kelurahan").Find(&posbankums)

	c.HTML(http.StatusOK, "paralegal_edit.html", gin.H{
		"Title":      "Edit Paralegal",
		"Paralegal":  paralegal,
		"Posbankums": posbankums,
	})
}

// ================== UPDATE ==================
func ParalegalUpdate(c *gin.Context) {
	id := c.Param("id")
	var paralegal models.Paralegal
	if err := config.DB.First(&paralegal, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	paralegal.Nama = c.PostForm("nama")
	posbankumID, _ := strconv.Atoi(c.PostForm("posbankum_id"))
	paralegal.PosbankumID = uint(posbankumID)

	file, err := c.FormFile("dokumen")
	if err == nil {
		if file.Size > 10*1024*1024 {
			c.HTML(http.StatusBadRequest, "paralegal_edit.html", gin.H{
				"Title":     "Edit Paralegal",
				"Paralegal": paralegal,
				"ErrorFile": "❌ Ukuran file maksimal 10MB",
			})
			return
		}

		if file.Header.Get("Content-Type") != "application/pdf" {
			c.HTML(http.StatusBadRequest, "paralegal_edit.html", gin.H{
				"Title":     "Edit Paralegal",
				"Paralegal": paralegal,
				"ErrorFile": "❌ File harus berupa PDF",
			})
			return
		}

		f, _ := file.Open()
		defer f.Close()
		fileBytes := make([]byte, file.Size)
		f.Read(fileBytes)

		paralegal.Dokumen = fileBytes
		paralegal.ContentType = file.Header.Get("Content-Type")
	}

	if err := config.DB.Save(&paralegal).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal update data")
		return
	}

	c.Redirect(http.StatusFound, "/jadi/admin/paralegal")
}

// ================== DELETE ==================
func ParalegalDelete(c *gin.Context) {
	id := c.Param("id")
	var paralegal models.Paralegal

	if err := config.DB.First(&paralegal, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	if err := config.DB.Delete(&paralegal).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal hapus data")
		return
	}

	c.Redirect(http.StatusFound, "/jadi/admin/paralegal")
}
