package controllers

import (
	"net/http"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-gonic/gin"
)

func LandingPage(c *gin.Context) {
	db := config.DB

	var totalPosbankum int64
	var totalKadarkum int64
	var totalPja int64
	var totalParalegal int64

	// Hitung total masing-masing tabel
	db.Model(&models.Posbankum{}).Count(&totalPosbankum)
	db.Model(&models.Kadarkum{}).Count(&totalKadarkum)
	db.Model(&models.PJA{}).Count(&totalPja)
	db.Model(&models.Paralegal{}).Count(&totalParalegal)

	// Kirim ke template
	c.HTML(http.StatusOK, "landing.html", gin.H{
		"TotalPosbankum": totalPosbankum,
		"TotalKadarkum":  totalKadarkum,
		"TotalPja":       totalPja,
		"TotalParalegal": totalParalegal,
	})
}
