package controllers

import (
	"math"
	"net/http"
	"strconv"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AdminPanel(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	if user == nil {
		c.Redirect(http.StatusFound, "/jadi/login")
		return
	}

	// Ambil query pencarian, kategori, dan pagination dari URL
	q := c.Query("q")
	selectedCategory := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	// Hitung data dari tabel untuk statistik dashboard
	var totalProvinsi, totalKabupaten, totalKecamatan, totalKelurahan int64
	var totalParalegal, totalPosbankum, totalPJA, totalKadarkum int64

	config.DB.Table("provinsis").Count(&totalProvinsi)
	config.DB.Table("kabupatens").Count(&totalKabupaten)
	config.DB.Table("kecamatans").Count(&totalKecamatan)
	config.DB.Table("kelurahans").Count(&totalKelurahan)
	config.DB.Table("paralegals").Count(&totalParalegal)
	config.DB.Table("posbankums").Count(&totalPosbankum)
	config.DB.Table("pjas").Count(&totalPJA)
	config.DB.Table("kadarkums").Count(&totalKadarkum)

	// Siapkan slice untuk hasil pencarian. Gunakan interface{} agar bisa menampung
	// slice dari berbagai model.
	var searchResults []interface{}
	var totalPages int

	// Lakukan pencarian hanya jika ada query dan kategori dipilih
	if q != "" && selectedCategory != "" {
		switch selectedCategory {
		case "posbankum":
			var results []models.Posbankum
			dbPos := config.DB.Preload("Kelurahan.Kecamatan.Kabupaten").
				Joins("JOIN kelurahans ON kelurahans.id = posbankums.kelurahan_id").
				Joins("JOIN kecamatans ON kecamatans.id = kelurahans.kecamatan_id").
				Joins("JOIN kabupatens ON kabupatens.id = kecamatans.kabupaten_id").
				Where("kelurahans.name LIKE ? OR kecamatans.name LIKE ? OR kabupatens.name LIKE ?",
					"%"+q+"%", "%"+q+"%", "%"+q+"%")

			var count int64
			dbPos.Model(&models.Posbankum{}).Count(&count)
			dbPos.Limit(limit).Offset(offset).Find(&results)
			totalPages = int(math.Ceil(float64(count) / float64(limit)))

			// Konversi hasil ke slice of interface{}
			for _, r := range results {
				searchResults = append(searchResults, r)
			}

		case "paralegal":
			var results []models.Paralegal
			dbPar := config.DB.Preload("Posbankum.Kelurahan.Kecamatan.Kabupaten").
				Joins("JOIN posbankums ON posbankums.id = paralegals.posbankum_id").
				Joins("JOIN kelurahans ON kelurahans.id = posbankums.kelurahan_id").
				Joins("JOIN kecamatans ON kecamatans.id = kelurahans.kecamatan_id").
				Joins("JOIN kabupatens ON kabupatens.id = kecamatans.kabupaten_id").
				Where("paralegals.nama LIKE ? OR kelurahans.name LIKE ? OR kecamatans.name LIKE ? OR kabupatens.name LIKE ?",
					"%"+q+"%", "%"+q+"%", "%"+q+"%", "%"+q+"%")

			var count int64
			dbPar.Model(&models.Paralegal{}).Count(&count)
			dbPar.Limit(limit).Offset(offset).Find(&results)
			totalPages = int(math.Ceil(float64(count) / float64(limit)))

			for _, r := range results {
				searchResults = append(searchResults, r)
			}

		case "kadarkum":
			var results []models.Kadarkum
			dbKad := config.DB.Preload("Kelurahan.Kecamatan.Kabupaten").
				Joins("JOIN kelurahans ON kelurahans.id = kadarkums.kelurahan_id").
				Joins("JOIN kecamatans ON kecamatans.id = kelurahans.kecamatan_id").
				Joins("JOIN kabupatens ON kabupatens.id = kecamatans.kabupaten_id").
				Where("kelurahans.name LIKE ? OR kecamatans.name LIKE ? OR kabupatens.name LIKE ?",
					"%"+q+"%", "%"+q+"%", "%"+q+"%")

			var count int64
			dbKad.Model(&models.Kadarkum{}).Count(&count)
			dbKad.Limit(limit).Offset(offset).Find(&results)
			totalPages = int(math.Ceil(float64(count) / float64(limit)))

			for _, r := range results {
				searchResults = append(searchResults, r)
			}

		case "pja":
			var results []models.PJA
			dbPja := config.DB.Preload("Kelurahan.Kecamatan.Kabupaten").
				Joins("JOIN kelurahans ON kelurahans.id = pjas.kelurahan_id").
				Joins("JOIN kecamatans ON kecamatans.id = kelurahans.kecamatan_id").
				Joins("JOIN kabupatens ON kabupatens.id = kecamatans.kabupaten_id").
				Where("kelurahans.name LIKE ? OR kecamatans.name LIKE ? OR kabupatens.name LIKE ?",
					"%"+q+"%", "%"+q+"%", "%"+q+"%")

			var count int64
			dbPja.Model(&models.PJA{}).Count(&count)
			dbPja.Limit(limit).Offset(offset).Find(&results)
			totalPages = int(math.Ceil(float64(count) / float64(limit)))

			for _, r := range results {
				searchResults = append(searchResults, r)
			}
		}
	}

	// Render ke admin.html dengan data yang sudah disiapkan
	c.HTML(http.StatusOK, "admin.html", gin.H{
		"Title":            "Dashboard",
		"user":             user,
		"Query":            q,
		"SelectedCategory": selectedCategory,
		"Page":             page,
		"Limit":            limit,
		"SearchResults":    searchResults,
		"TotalPages":       totalPages,
		"totalProvinsi":    totalProvinsi,
		"totalKabupaten":   totalKabupaten,
		"totalKecamatan":   totalKecamatan,
		"totalKelurahan":   totalKelurahan,
		"totalParalegal":   totalParalegal,
		"totalPosbankum":   totalPosbankum,
		"totalPJA":         totalPJA,
		"totalKadarkum":    totalKadarkum,
	})
}
