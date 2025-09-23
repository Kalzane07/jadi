package controllers

import (
	"go-admin/config"
	"go-admin/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// KelurahanSearch menyediakan data untuk fitur pencarian autocomplete.
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
