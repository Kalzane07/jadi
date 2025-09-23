package controllers

import (
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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

	var pjas []models.PJA
	db := config.DB.Model(&models.PJA{}).
		Preload("Kelurahan").
		Preload("Kelurahan.Kecamatan").
		Preload("Kelurahan.Kecamatan.Kabupaten")

	if search != "" {
		db = db.Joins("JOIN kelurahans ON kelurahans.id = pjas.kelurahan_id").
			Where("kelurahans.name LIKE ? OR pjas.catatan LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		log.Println("Error menghitung total data:", err)
		c.String(http.StatusInternalServerError, "Error menghitung total data")
		return
	}

	if err := db.Offset(offset).Limit(limit).Find(&pjas).Error; err != nil {
		log.Println("Error mengambil data:", err)
		c.String(http.StatusInternalServerError, "Error mengambil data")
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	session := sessions.Default(c)
	user := session.Get("user")

	c.HTML(http.StatusOK, "pja_index.html", gin.H{
		"Title":      "Data PJA",
		"Pjas":       pjas,
		"Search":     search,
		"Page":       page,
		"Offset":     offset,
		"TotalPages": totalPages,
		"user":       user,
	})
}

// ================== CREATE FORM ==================
func PJACreate(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	c.HTML(http.StatusOK, "pja_create.html", gin.H{
		"Title": "Tambah PJA",
		"user":  user,
	})
}

// ================== STORE ==================
func PJAStore(c *gin.Context) {
	kelurahanID, _ := strconv.Atoi(c.PostForm("kelurahan_id"))
	catatan := c.PostForm("catatan")

	// cek duplikat
	var existing models.PJA
	if err := config.DB.Where("kelurahan_id = ?", kelurahanID).First(&existing).Error; err == nil {
		c.HTML(http.StatusBadRequest, "pja_create.html", gin.H{
			"Title":          "Tambah PJA",
			"ErrorKelurahan": "❌ PJA untuk kelurahan ini sudah ada",
			"Catatan":        catatan,
		})
		return
	}

	file, err := c.FormFile("dokumen")
	if err != nil {
		c.HTML(http.StatusBadRequest, "pja_create.html", gin.H{
			"Title":     "Tambah PJA",
			"ErrorFile": "❌ Dokumen wajib diupload",
			"Catatan":   catatan,
		})
		return
	}

	if file.Size > 10*1024*1024 { // max 10MB
		c.HTML(http.StatusBadRequest, "pja_create.html", gin.H{
			"Title":     "Tambah PJA",
			"ErrorFile": "❌ Ukuran file maksimal 10MB",
			"Catatan":   catatan,
		})
		return
	}

	if strings.ToLower(filepath.Ext(file.Filename)) != ".pdf" {
		c.HTML(http.StatusBadRequest, "pja_create.html", gin.H{
			"Title":     "Tambah PJA",
			"ErrorFile": "❌ File harus berupa PDF",
			"Catatan":   catatan,
		})
		return
	}

	// baca file ke []byte
	fileContent, err := file.Open()
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal membuka file")
		return
	}
	defer fileContent.Close()

	data := make([]byte, file.Size)
	_, err = fileContent.Read(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal membaca file")
		return
	}

	pja := models.PJA{
		KelurahanID: uint(kelurahanID),
		Dokumen:     data,
		Catatan:     catatan,
	}

	if err := config.DB.Create(&pja).Error; err != nil {
		log.Println("Gagal menyimpan data ke database:", err)
		c.String(http.StatusInternalServerError, "Gagal menyimpan data ke database")
		return
	}

	c.Redirect(http.StatusFound, "/jadi/admin/pja")
}

// ================== EDIT FORM ==================
func PJAEdit(c *gin.Context) {
	id := c.Param("id")
	var pja models.PJA
	if err := config.DB.
		Preload("Kelurahan").
		Preload("Kelurahan.Kecamatan").
		Preload("Kelurahan.Kecamatan.Kabupaten").
		First(&pja, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "Data tidak ditemukan")
		} else {
			c.String(http.StatusInternalServerError, "Error database")
		}
		return
	}

	session := sessions.Default(c)
	user := session.Get("user")

	c.HTML(http.StatusOK, "pja_edit.html", gin.H{
		"Title": "Edit PJA",
		"PJA":   pja,
		"user":  user,
	})
}

// ================== UPDATE ==================
func PJAUpdate(c *gin.Context) {
	id := c.Param("id")
	var pja models.PJA
	if err := config.DB.First(&pja, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "Data tidak ditemukan")
		} else {
			c.String(http.StatusInternalServerError, "Error database")
		}
		return
	}

	kelurahanID, _ := strconv.Atoi(c.PostForm("kelurahan_id"))
	catatan := c.PostForm("catatan")

	// cek duplikat kelurahan lain
	var count int64
	config.DB.Model(&models.PJA{}).
		Where("kelurahan_id = ? AND id <> ?", kelurahanID, pja.ID).
		Count(&count)
	if count > 0 {
		session := sessions.Default(c)
		user := session.Get("user")
		c.HTML(http.StatusBadRequest, "pja_edit.html", gin.H{
			"Title":          "Edit PJA",
			"PJA":            pja,
			"ErrorKelurahan": "❌ PJA untuk kelurahan ini sudah ada",
			"user":           user,
		})
		return
	}

	pja.KelurahanID = uint(kelurahanID)
	pja.Catatan = catatan

	// jika ada file baru
	file, err := c.FormFile("dokumen")
	if err == nil {
		if file.Size > 10*1024*1024 {
			c.String(http.StatusBadRequest, "❌ Ukuran file maksimal 10MB")
			return
		}
		if strings.ToLower(filepath.Ext(file.Filename)) != ".pdf" {
			c.String(http.StatusBadRequest, "❌ File harus berupa PDF")
			return
		}

		fileContent, err := file.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, "Gagal membuka file baru")
			return
		}
		defer fileContent.Close()

		data := make([]byte, file.Size)
		_, err = fileContent.Read(data)
		if err != nil {
			c.String(http.StatusInternalServerError, "Gagal membaca file baru")
			return
		}
		pja.Dokumen = data
	}

	if err := config.DB.Save(&pja).Error; err != nil {
		log.Println("Gagal mengupdate data di database:", err)
		c.String(http.StatusInternalServerError, "Gagal mengupdate data di database")
		return
	}
	c.Redirect(http.StatusFound, "/jadi/admin/pja")
}

// ================== DELETE ==================
func PJADelete(c *gin.Context) {
	id := c.Param("id")
	var pja models.PJA

	if err := config.DB.First(&pja, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "Data tidak ditemukan")
		} else {
			log.Println("Error database:", err)
			c.String(http.StatusInternalServerError, "Error database")
		}
		return
	}

	// hapus record database (tidak ada file di disk)
	if err := config.DB.Delete(&pja).Error; err != nil {
		log.Println("Gagal menghapus data dari database:", err)
		c.String(http.StatusInternalServerError, "Gagal menghapus data dari database")
		return
	}

	c.Redirect(http.StatusFound, "/jadi/admin/pja")
}
