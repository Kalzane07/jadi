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

// ================== API SEARCH POSBANKUM ==================
func PosbankumSearch(c *gin.Context) {
	term := c.Query("term")

	var posbankums []struct {
		ID        uint   `json:"id"`
		Text      string `json:"text"`
		Kelurahan string `json:"kelurahan"`
		Kecamatan string `json:"kecamatan"`
		Kabupaten string `json:"kabupaten"`
	}

	query := config.DB.Table("posbankums").
		Select(`
			posbankums.id,
			CONCAT(kelurahans.name, " - ", kecamatans.name, " - ", kabupatens.name) as text,
			kelurahans.name as kelurahan,
			kecamatans.name as kecamatan,
			kabupatens.name as kabupaten`).
		Joins("JOIN kelurahans ON kelurahans.id = posbankums.kelurahan_id").
		Joins("JOIN kecamatans ON kecamatans.id = kelurahans.kecamatan_id").
		Joins("JOIN kabupatens ON kabupatens.id = kecamatans.kabupaten_id")

	if term != "" {
		query = query.Where("kelurahans.name LIKE ? OR kecamatans.name LIKE ? OR kabupatens.name LIKE ?",
			"%"+term+"%", "%"+term+"%", "%"+term+"%")
	}

	query.Limit(20).Scan(&posbankums)

	c.JSON(http.StatusOK, gin.H{
		"results": posbankums,
	})
}

// ================== STORE ==================
func ParalegalStore(c *gin.Context) {
	posbankumID, _ := strconv.Atoi(c.PostForm("posbankum_id"))

	nama := utils.SanitizeInput(c.PostForm("nama"))
	var dokumenPath string

	file, err := c.FormFile("dokumen")
	if err == nil {
		if !utils.ValidatePDFUpload(c, file) {
			c.HTML(http.StatusOK, "paralegal_create.html", gin.H{
				"Title":     "Tambah Paralegal",
				"ErrorFile": "❌ File tidak valid. Pastikan file adalah PDF dan ukurannya di bawah 10MB.",
				"Nama":      nama,
			})
			return
		}

		uploadPath := "uploads/paralegal"
		os.MkdirAll(uploadPath, os.ModePerm)

		// generate nama file unik menggunakan uuid
		ext := filepath.Ext(file.Filename)
		newName := uuid.New().String() + ext
		fullPath := filepath.Join(uploadPath, newName)

		if err := c.SaveUploadedFile(file, fullPath); err != nil {
			c.HTML(http.StatusOK, "paralegal_create.html", gin.H{
				"Title":     "Tambah Paralegal",
				"ErrorFile": "Gagal upload file",
				"Nama":      nama,
			})
			return
		}
		dokumenPath = strings.ReplaceAll(fullPath, "\\", "/")
	}

	paralegal := models.Paralegal{
		PosbankumID: uint(posbankumID),
		Nama:        nama,
		Dokumen:     dokumenPath,
	}

	config.DB.Create(&paralegal)
	c.Redirect(http.StatusFound, "/admin/paralegal")
}

func ParalegalView(c *gin.Context) {
	id := c.Param("id")
	var paralegal models.Paralegal
	if err := config.DB.
		Preload("Posbankum").
		Preload("Posbankum.Kelurahan").
		Preload("Posbankum.Kelurahan.Kecamatan").
		Preload("Posbankum.Kelurahan.Kecamatan.Kabupaten").
		First(&paralegal, id).Error; err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}
	c.Header("Content-Disposition", "inline; filename="+filepath.Base(paralegal.Dokumen))
	c.Header("Content-Type", "application/pdf")
	c.File(paralegal.Dokumen)
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

	paralegal.Nama = utils.SanitizeInput(c.PostForm("nama"))
	posbankumID, _ := strconv.Atoi(c.PostForm("posbankum_id"))
	paralegal.PosbankumID = uint(posbankumID)

	file, err := c.FormFile("dokumen")
	if err == nil {
		if !utils.ValidatePDFUpload(c, file) {
			c.HTML(http.StatusOK, "paralegal_edit.html", gin.H{
				"Title":     "Edit Paralegal",
				"Paralegal": paralegal,
				"ErrorFile": "❌ File tidak valid. Pastikan file adalah PDF dan ukurannya di bawah 10MB.",
			})
			return
		}

		uploadPath := "uploads/paralegal"
		os.MkdirAll(uploadPath, os.ModePerm)

		// generate nama file unik menggunakan uuid
		ext := filepath.Ext(file.Filename)
		newName := uuid.New().String() + ext
		newPath := filepath.Join(uploadPath, newName)

		if err := c.SaveUploadedFile(file, newPath); err != nil {
			c.HTML(http.StatusOK, "paralegal_edit.html", gin.H{
				"Title":     "Edit Paralegal",
				"Paralegal": paralegal,
				"ErrorFile": "❌ Gagal upload file",
			})
			return
		}

		// hapus file lama kalau ada
		if paralegal.Dokumen != "" {
			_ = os.Remove(paralegal.Dokumen)
		}

		paralegal.Dokumen = strings.ReplaceAll(newPath, "\\", "/")
	}

	config.DB.Save(&paralegal)
	c.Redirect(http.StatusFound, "/admin/paralegal")
}

// ================== DELETE ==================
func ParalegalDelete(c *gin.Context) {
	id := c.Param("id")
	var paralegal models.Paralegal

	if err := config.DB.First(&paralegal, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	// hapus file PDF kalau ada
	if paralegal.Dokumen != "" {
		_ = os.Remove(paralegal.Dokumen)
	}

	// hapus record
	config.DB.Delete(&paralegal)

	c.Redirect(http.StatusFound, "/admin/paralegal")
}
