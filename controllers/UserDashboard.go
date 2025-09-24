package controllers

import (
	"mime"
	"net/http"
	"path/filepath"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-gonic/gin"
)

// ==================== STRUCT BARU UNTUK PARALEGAL ====================

// Struktur untuk menampung nama dan dokumen Paralegal
type ParalegalData struct {
	ID      uint // ID diperlukan untuk link dokumen di template
	Nama    string
	Dokumen string
}

// Struktur untuk data Paralegal per kelurahan
type KelurahanParalegal struct {
	NamaKelurahan string
	Total         int
	Paralegals    []ParalegalData
}

// Struktur untuk data Paralegal per kecamatan
type KecamatanParalegal struct {
	NamaKecamatan string
	Total         int
	Kelurahans    []KelurahanParalegal
}

// Struktur untuk data Paralegal per kabupaten
type KabupatenParalegal struct {
	NamaKabupaten string
	Total         int
	Kecamatans    []KecamatanParalegal
}

// ==================== STRUCT LAINNYA ====================

// Struktur buat nampung data dokumen per kelurahan
type KelurahanDokumen struct {
	NamaKelurahan string
	AdaDokumen    bool
	Total         int
	Tercapai      int
	Persentase    float64
	Posbankums    []models.Posbankum
	Kadarkums     []models.Kadarkum
	Pjas          []models.Pja
	Paralegals    []models.Paralegal // Field ini sepertinya tidak terpakai di sini, tapi tidak apa-apa
}

// Struktur buat nampung data summary per kecamatan
type KecamatanSummary struct {
	NamaKecamatan string
	Total         int
	Tercapai      int
	Persentase    float64
	Kelurahans    []KelurahanDokumen
}

// Struktur buat nampung data summary per kabupaten
type KabupatenSummary struct {
	NamaKabupaten string
	Total         int
	Tercapai      int
	Persentase    float64
	Kecamatans    []KecamatanSummary
}

// Struktur buat data dashboard user
type DashboardData struct {
	Title                   string
	Provinsi                string
	Posbankum               []KabupatenSummary
	Kadarkum                []KabupatenSummary
	PJA                     []KabupatenSummary
	Paralegal               []KabupatenParalegal
	TotalPosbankumProvinsi  int
	TotalKadarkumProvinsi   int
	TotalPjaProvinsi        int
	TotalParalegalProvinsi  int
	TotalKelurahanProvinsi  int
	PersenPosbankumProvinsi float64
	PersenKadarkumProvinsi  float64
	PersenPjaProvinsi       float64
	BaseHref                string
}

// helper hitung persentase aman
func hitungPersen(tercapai, total int) float64 {
	if total == 0 {
		return 0
	}
	return (float64(tercapai) / float64(total)) * 100
}

// ==================== CONTROLLER ====================

func UserDashboard(c *gin.Context) {
	var provinsi models.Provinsi

	if err := config.DB.Preload("Kabupatens.Kecamatans.Kelurahans").First(&provinsi).Error; err != nil {
		c.String(http.StatusInternalServerError, "❌ Tidak ada provinsi di database")
		return
	}

	var (
		posbankumAll, kadarkumAll, pjaAll []KabupatenSummary
		paralegalAll                      []KabupatenParalegal
		totalPosProv, tercapaiPosProv     int
		totalKadProv, tercapaiKadProv     int
		totalPJAProv, tercapaiPJAProv     int
		totalParalegalProv                int
		totalKelurahanProv64              int64
	)

	config.DB.Model(&models.Kelurahan{}).Count(&totalKelurahanProv64)
	totalKelurahanProv := int(totalKelurahanProv64)

	for _, kab := range provinsi.Kabupatens {
		var posbankumKec, kadarkumKec, pjaKec []KecamatanSummary
		var paralegalKec []KecamatanParalegal

		totalPosbankumKab, tercapaiPosbankumKab := 0, 0
		totalKadarkumKab, tercapaiKadarkumKab := 0, 0
		totalPJAKab, tercapaiPJAKab := 0, 0
		totalParalegalKab := 0

		for _, kec := range kab.Kecamatans {
			// ================== POSBANKUM ==================
			var totalPos, tercapaiPos int64
			config.DB.Model(&models.Kelurahan{}).Where("kecamatan_id = ?", kec.ID).Count(&totalPos)
			config.DB.Model(&models.Posbankum{}).
				Joins("JOIN kelurahans ON kelurahans.id = posbankums.kelurahan_id").
				Where("kelurahans.kecamatan_id = ?", kec.ID).Count(&tercapaiPos)

			var kelurahanDocsPos []KelurahanDokumen
			for _, kel := range kec.Kelurahans {
				var posbankums []models.Posbankum
				config.DB.Where("kelurahan_id = ?", kel.ID).Find(&posbankums)

				tercapai := 0
				if len(posbankums) > 0 {
					tercapai = 1
				}

				kelurahanDocsPos = append(kelurahanDocsPos, KelurahanDokumen{
					NamaKelurahan: kel.Name,
					Total:         1,
					Tercapai:      tercapai,
					Persentase:    hitungPersen(tercapai, 1),
					Posbankums:    posbankums,
				})
			}

			posbankumKec = append(posbankumKec, KecamatanSummary{
				NamaKecamatan: kec.Name,
				Total:         int(totalPos),
				Tercapai:      int(tercapaiPos),
				Persentase:    hitungPersen(int(tercapaiPos), int(totalPos)),
				Kelurahans:    kelurahanDocsPos,
			})
			totalPosbankumKab += int(totalPos)
			tercapaiPosbankumKab += int(tercapaiPos)

			// ================== KADARKUM ==================
			var totalK, tercapaiK int64
			config.DB.Model(&models.Kelurahan{}).Where("kecamatan_id = ?", kec.ID).Count(&totalK)
			config.DB.Model(&models.Kadarkum{}).
				Joins("JOIN kelurahans ON kelurahans.id = kadarkums.kelurahan_id").
				Where("kelurahans.kecamatan_id = ?", kec.ID).Count(&tercapaiK)

			var kelurahanDocsKadarkum []KelurahanDokumen
			for _, kel := range kec.Kelurahans {
				var kadarkums []models.Kadarkum
				config.DB.Where("kelurahan_id = ?", kel.ID).Find(&kadarkums)
				tercapai := 0
				if len(kadarkums) > 0 {
					tercapai = 1
				}
				kelurahanDocsKadarkum = append(kelurahanDocsKadarkum, KelurahanDokumen{
					NamaKelurahan: kel.Name,
					Total:         1,
					Tercapai:      tercapai,
					Persentase:    hitungPersen(tercapai, 1),
					Kadarkums:     kadarkums,
				})
			}

			kadarkumKec = append(kadarkumKec, KecamatanSummary{
				NamaKecamatan: kec.Name,
				Total:         int(totalK),
				Tercapai:      int(tercapaiK),
				Persentase:    hitungPersen(int(tercapaiK), int(totalK)),
				Kelurahans:    kelurahanDocsKadarkum,
			})
			totalKadarkumKab += int(totalK)
			tercapaiKadarkumKab += int(tercapaiK)

			// ================== PJA ==================
			var totalP, tercapaiP int64
			config.DB.Model(&models.Kelurahan{}).Where("kecamatan_id = ?", kec.ID).Count(&totalP)
			config.DB.Model(&models.Pja{}).
				Joins("JOIN kelurahans ON kelurahans.id = pjas.kelurahan_id").
				Where("kelurahans.kecamatan_id = ?", kec.ID).Count(&tercapaiP)

			var kelurahanDocsPja []KelurahanDokumen
			for _, kel := range kec.Kelurahans {
				var pjas []models.Pja
				config.DB.Where("kelurahan_id = ?", kel.ID).Find(&pjas)
				tercapai := 0
				if len(pjas) > 0 {
					tercapai = 1
				}
				kelurahanDocsPja = append(kelurahanDocsPja, KelurahanDokumen{
					NamaKelurahan: kel.Name,
					Total:         1,
					Tercapai:      tercapai,
					Persentase:    hitungPersen(tercapai, 1),
					Pjas:          pjas,
				})
			}

			pjaKec = append(pjaKec, KecamatanSummary{
				NamaKecamatan: kec.Name,
				Total:         int(totalP),
				Tercapai:      int(tercapaiP),
				Persentase:    hitungPersen(int(tercapaiP), int(totalP)),
				Kelurahans:    kelurahanDocsPja,
			})
			totalPJAKab += int(totalP)
			tercapaiPJAKab += int(tercapaiP)

			// ================== PARALEGAL ==================
			var kelurahanParalegals []KelurahanParalegal
			totalParalegalKec := 0

			for _, kel := range kec.Kelurahans {
				var paralegalsFromDB []models.Paralegal
				config.DB.Model(&models.Paralegal{}).
					Joins("JOIN posbankums ON posbankums.id = paralegals.posbankum_id").
					Where("posbankums.kelurahan_id = ?", kel.ID).
					Find(&paralegalsFromDB)

				var paralegalDataForKelurahan []ParalegalData
				for _, p := range paralegalsFromDB {
					paralegalDataForKelurahan = append(paralegalDataForKelurahan, ParalegalData{
						ID:      p.ID, // Kirim ID ke template
						Nama:    p.Nama,
						Dokumen: p.Dokumen,
					})
				}

				kelurahanParalegals = append(kelurahanParalegals, KelurahanParalegal{
					NamaKelurahan: kel.Name,
					Total:         len(paralegalsFromDB),
					Paralegals:    paralegalDataForKelurahan,
				})
				totalParalegalKec += len(paralegalsFromDB)
			}
			// KURUNG KURAWAL YANG SALAH SEBELUMNYA ADA DI SINI. SUDAH DIHAPUS.

			paralegalKec = append(paralegalKec, KecamatanParalegal{
				NamaKecamatan: kec.Name,
				Total:         totalParalegalKec,
				Kelurahans:    kelurahanParalegals,
			})
			totalParalegalKab += totalParalegalKec

		} // <-- KURUNG KURAWAL DIPINDAHKAN KE SINI UNTUK MENUTUP LOOP KECAMATAN DENGAN BENAR

		// Push ke level kabupaten
		posbankumAll = append(posbankumAll, KabupatenSummary{
			NamaKabupaten: kab.Name,
			Total:         totalPosbankumKab,
			Tercapai:      tercapaiPosbankumKab,
			Persentase:    hitungPersen(tercapaiPosbankumKab, totalPosbankumKab),
			Kecamatans:    posbankumKec,
		})

		kadarkumAll = append(kadarkumAll, KabupatenSummary{
			NamaKabupaten: kab.Name,
			Total:         totalKadarkumKab,
			Tercapai:      tercapaiKadarkumKab,
			Persentase:    hitungPersen(tercapaiKadarkumKab, totalKadarkumKab),
			Kecamatans:    kadarkumKec,
		})

		pjaAll = append(pjaAll, KabupatenSummary{
			NamaKabupaten: kab.Name,
			Total:         totalPJAKab,
			Tercapai:      tercapaiPJAKab,
			Persentase:    hitungPersen(tercapaiPJAKab, totalPJAKab),
			Kecamatans:    pjaKec,
		})

		paralegalAll = append(paralegalAll, KabupatenParalegal{
			NamaKabupaten: kab.Name,
			Total:         totalParalegalKab,
			Kecamatans:    paralegalKec,
		})

		// ====== Akumulasi ke provinsi ======
		totalPosProv += totalPosbankumKab
		tercapaiPosProv += tercapaiPosbankumKab
		totalKadProv += totalKadarkumKab
		tercapaiKadProv += tercapaiKadarkumKab
		totalPJAProv += totalPJAKab
		tercapaiPJAProv += tercapaiPJAKab
		totalParalegalProv += totalParalegalKab
	}

	data := DashboardData{
		Title:                   "Dashboard User",
		Provinsi:                provinsi.Name,
		Posbankum:               posbankumAll,
		Kadarkum:                kadarkumAll,
		PJA:                     pjaAll,
		Paralegal:               paralegalAll,
		TotalPosbankumProvinsi:  tercapaiPosProv,
		TotalKadarkumProvinsi:   tercapaiKadProv,
		TotalPjaProvinsi:        tercapaiPJAProv,
		TotalParalegalProvinsi:  totalParalegalProv,
		TotalKelurahanProvinsi:  totalKelurahanProv,
		PersenPosbankumProvinsi: hitungPersen(tercapaiPosProv, totalKelurahanProv),
		PersenKadarkumProvinsi:  hitungPersen(tercapaiKadProv, totalKelurahanProv),
		PersenPjaProvinsi:       hitungPersen(tercapaiPJAProv, totalKelurahanProv),
		BaseHref:                "/jadi",
	}

	c.HTML(http.StatusOK, "user_dashboard.html", data)
}

// ViewDocument adalah handler universal untuk menampilkan dokumen
func ViewDocument(c *gin.Context) {
	docType := c.Param("type")
	id := c.Param("id")
	var filePath string

	switch docType {
	case "posbankum":
		var data models.Posbankum
		if err := config.DB.First(&data, id).Error; err != nil {
			c.String(http.StatusNotFound, "Dokumen Posbankum tidak ditemukan")
			return
		}
		filePath = data.Dokumen
	case "paralegal":
		var data models.Paralegal
		if err := config.DB.First(&data, id).Error; err != nil {
			c.String(http.StatusNotFound, "Dokumen Paralegal tidak ditemukan")
			return
		}
		filePath = data.Dokumen
	case "pja":
		var data models.Pja
		if err := config.DB.First(&data, id).Error; err != nil {
			c.String(http.StatusNotFound, "Dokumen PJA tidak ditemukan")
			return
		}
		filePath = data.Dokumen
	case "kadarkum":
		var data models.Kadarkum
		if err := config.DB.First(&data, id).Error; err != nil {
			c.String(http.StatusNotFound, "Dokumen Kadarkum tidak ditemukan")
			return
		}
		filePath = data.Dokumen
	default:
		c.String(http.StatusBadRequest, "Tipe dokumen tidak valid")
		return
	}

	if filePath == "" {
		c.String(http.StatusNotFound, "Path dokumen kosong atau tidak tersedia.")
		return
	}

	fileName := filepath.Base(filePath)
	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Disposition", "inline; filename="+fileName)
	c.Header("Content-Type", contentType)
	c.File(filePath)
}
