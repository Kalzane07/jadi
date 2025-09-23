package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-gonic/gin"
)

// ================== GENERIC FILE SYSTEM (untuk data lama) ==================
func ServeDokumenFile(c *gin.Context) {
	tipe := c.Param("tipe")
	filename := c.Param("filename")

	// Lokasi folder upload
	filePath := filepath.Join("uploads", tipe, filename)

	// Cek file ada atau tidak
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "File tidak ditemukan")
		return
	}

	c.File(filePath)
}

//
// ================== DOKUMEN DARI DATABASE (per modul) ==================
//

// ViewPosbankumDokumen
func ViewPosbankumDokumen(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "ID tidak valid")
		return
	}

	var posbankum models.Posbankum
	if err := config.DB.First(&posbankum, id).Error; err != nil {
		c.String(http.StatusNotFound, "Dokumen tidak ditemukan")
		return
	}

	if posbankum.Dokumen != nil {
		c.Header("Content-Type", posbankum.ContentType)
		c.Data(http.StatusOK, posbankum.ContentType, posbankum.Dokumen)
	} else {
		c.String(http.StatusNotFound, "Dokumen kosong")
	}
}

// ViewParalegalDokumen
func ViewParalegalDokumen(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "ID tidak valid")
		return
	}

	var paralegal models.Paralegal
	if err := config.DB.First(&paralegal, id).Error; err != nil {
		c.String(http.StatusNotFound, "Dokumen tidak ditemukan")
		return
	}

	if paralegal.Dokumen != nil {
		c.Header("Content-Type", paralegal.ContentType)
		c.Data(http.StatusOK, paralegal.ContentType, paralegal.Dokumen)
	} else {
		c.String(http.StatusNotFound, "Dokumen kosong")
	}
}

// ViewPJADokumen
func ViewPJADokumen(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "ID tidak valid")
		return
	}

	var pja models.PJA
	if err := config.DB.First(&pja, id).Error; err != nil {
		c.String(http.StatusNotFound, "Dokumen tidak ditemukan")
		return
	}

	if pja.Dokumen != nil {
		c.Header("Content-Type", pja.ContentType)
		c.Data(http.StatusOK, pja.ContentType, pja.Dokumen)
	} else {
		c.String(http.StatusNotFound, "Dokumen kosong")
	}
}

// ViewKadarkumDokumen
func ViewKadarkumDokumen(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "ID tidak valid")
		return
	}

	var kadarkum models.Kadarkum
	if err := config.DB.First(&kadarkum, id).Error; err != nil {
		c.String(http.StatusNotFound, "Dokumen tidak ditemukan")
		return
	}

	if kadarkum.Dokumen != nil {
		c.Header("Content-Type", kadarkum.ContentType)
		c.Data(http.StatusOK, kadarkum.ContentType, kadarkum.Dokumen)
	} else {
		c.String(http.StatusNotFound, "Dokumen kosong")
	}
}
