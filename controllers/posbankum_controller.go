package controllers

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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

// ================== STORE DOKUMEN KE DATABASE ==================
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

	fileType := file.Header.Get("Content-Type")
	if fileType != "application/pdf" {
		c.HTML(http.StatusBadRequest, "posbankum_create.html", gin.H{
			"Title":     "Tambah Posbankum",
			"ErrorFile": "File harus berformat PDF.",
			"Catatan":   catatan,
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		log.Println("Gagal memproses file:", err)
		c.String(http.StatusInternalServerError, "Gagal memproses file")
		return
	}
	defer src.Close()

	fileBytes, err := ioutil.ReadAll(src)
	if err != nil {
		log.Println("Gagal membaca file:", err)
		c.String(http.StatusInternalServerError, "Gagal membaca file")
		return
	}

	posbankum := models.Posbankum{
		KelurahanID: uint(kelurahanID),
		Dokumen:     fileBytes,
		ContentType: fileType,
		Catatan:     catatan,
	}

	if err := config.DB.Create(&posbankum).Error; err != nil {
		log.Println("Gagal menyimpan data ke database:", err)
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

// ================== UPDATE DOKUMEN DI DATABASE ==================
func PosbankumUpdate(c *gin.Context) {
	id := c.Param("id")
	var posbankum models.Posbankum

	if err := config.DB.First(&posbankum, id).Error; err != nil {
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
		fileType := file.Header.Get("Content-Type")
		if fileType != "application/pdf" {
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

		src, err := file.Open()
		if err != nil {
			log.Println("Gagal membuka file:", err)
			c.String(http.StatusInternalServerError, "Gagal memproses file baru")
			return
		}
		defer src.Close()

		fileBytes, err := ioutil.ReadAll(src)
		if err != nil {
			log.Println("Gagal membaca file:", err)
			c.String(http.StatusInternalServerError, "Gagal membaca file baru")
			return
		}

		// ✅ Update field dokumen dan tipe konten dengan data file baru
		posbankum.Dokumen = fileBytes
		posbankum.ContentType = fileType
	}

	if err := config.DB.Save(&posbankum).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal mengupdate data di database")
		return
	}
	c.Redirect(http.StatusFound, "/jadi/admin/posbankum")
}

// ================== DELETE DOKUMEN DARI DATABASE ==================
func PosbankumDelete(c *gin.Context) {
	id := c.Param("id")
	var posbankum models.Posbankum

	if err := config.DB.First(&posbankum, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	// ✅ Cukup hapus record dari database, tidak ada file yang perlu dihapus dari sistem file
	if err := config.DB.Delete(&posbankum).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal menghapus data dari database")
		return
	}

	c.Redirect(http.StatusFound, "/jadi/admin/posbankum")
}

// ================== API SEARCH ==================
func PosbankumSearch(c *gin.Context) {
	term := c.Query("term")

	var results []struct {
		ID        uint   `json:"id"`
		Kelurahan string `json:"kelurahan"`
		Kecamatan string `json:"kecamatan"`
		Kabupaten string `json:"kabupaten"`
	}

	// Join biar bisa dapet nama wilayah
	err := config.DB.Table("posbankums").
		Select("posbankums.id, kelurahans.name AS kelurahan, kecamatans.name AS kecamatan, kabupatens.name AS kabupaten").
		Joins("JOIN kelurahans ON kelurahans.id = posbankums.kelurahan_id").
		Joins("JOIN kecamatans ON kecamatans.id = kelurahans.kecamatan_id").
		Joins("JOIN kabupatens ON kabupatens.id = kecamatans.kabupaten_id").
		Where("kelurahans.name LIKE ? OR kecamatans.name LIKE ? OR kabupatens.name LIKE ?", "%"+term+"%", "%"+term+"%", "%"+term+"%").
		Limit(20).
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal ambil data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}
