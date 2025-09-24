package controllers

import (
	"net/http"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-gonic/gin"
)

func ProvinsiIndex(c *gin.Context) {
	var provinsis []models.Provinsi
	config.DB.Find(&provinsis)

	c.HTML(http.StatusOK, "provinsi_index.html", gin.H{
		"Title":     "Data Provinsi",
		"Provinsis": provinsis,
		"BaseHref":  "/jadi",
	})
}

func KabupatenIndex(c *gin.Context) {
	var kabupatens []models.Kabupaten
	config.DB.Preload("Provinsi").Find(&kabupatens)

	c.HTML(http.StatusOK, "kabupaten_index.html", gin.H{
		"Title":      "Data Kabupaten/Kota",
		"Kabupatens": kabupatens,
		"BaseHref":   "/jadi",
	})
}

func KecamatanIndex(c *gin.Context) {
	var kecamatans []models.Kecamatan
	config.DB.Preload("Kabupaten").Find(&kecamatans)

	c.HTML(http.StatusOK, "kecamatan_index.html", gin.H{
		"Title":      "Data Kecamatan",
		"Kecamatans": kecamatans,
		"BaseHref":   "/jadi",
	})
}

func KelurahanIndex(c *gin.Context) {
	var kelurahans []models.Kelurahan
	config.DB.Preload("Kecamatan").Find(&kelurahans)

	c.HTML(http.StatusOK, "kelurahan_index.html", gin.H{
		"Title":      "Data Kelurahan/Desa",
		"Kelurahans": kelurahans,
		"BaseHref":   "/jadi",
	})
}
