package controllers

import (
	"net/http"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-gonic/gin"
)

// ==================== STRUCT BARU UNTUK PARALEGAL ====================

// Struktur untuk menampung nama dan dokumen Paralegal
type ParalegalData struct {
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
	Dokumen       []string
	Total         int
	Tercapai      int
	Persentase    float64
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
	Paralegal               []KabupatenParalegal // PERUBAHAN TIPE DATA DI SINI
	TotalPosbankumProvinsi  int
	TotalKadarkumProvinsi   int
	TotalPjaProvinsi        int
	TotalParalegalProvinsi  int
	TotalKelurahanProvinsi  int
	PersenPosbankumProvinsi float64
	PersenKadarkumProvinsi  float64
	PersenPjaProvinsi       float64
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
		paralegalAll                      []KabupatenParalegal // PERUBAHAN TIPE DATA DI SINI
		totalPosProv, tercapaiPosProv     int
		totalKadProv, tercapaiKadProv     int
		totalPJAProv, tercapaiPJAProv     int
		totalParalegalProv                int
		totalKelurahanProv64              int64
	)

	// hitung total kelurahan di provinsi (pakai int64 -> aman buat GORM)
	config.DB.Model(&models.Kelurahan{}).Count(&totalKelurahanProv64)
	totalKelurahanProv := int(totalKelurahanProv64)

	for _, kab := range provinsi.Kabupatens {
		var posbankumKec, kadarkumKec, pjaKec []KecamatanSummary
		var paralegalKec []KecamatanParalegal // STRUCT BARU

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
				Where("kelurahans.kecamatan_id = ? AND posbankums.dokumen <> ''", kec.ID).Count(&tercapaiPos)

			var kelurahanDocsPos []KelurahanDokumen
			for _, kel := range kec.Kelurahans {
				var docs []models.Posbankum
				config.DB.Where("kelurahan_id = ?", kel.ID).Find(&docs)

				var dokList []string
				for _, d := range docs {
					if d.Dokumen != "" {
						dokList = append(dokList, d.Dokumen)
					}
				}

				tercapai := 0
				if len(dokList) > 0 {
					tercapai = 1
				}

				kelurahanDocsPos = append(kelurahanDocsPos, KelurahanDokumen{
					NamaKelurahan: kel.Name,
					AdaDokumen:    tercapai == 1,
					Dokumen:       dokList,
					Total:         1,
					Tercapai:      tercapai,
					Persentase:    hitungPersen(tercapai, 1),
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
				Where("kelurahans.kecamatan_id = ? AND kadarkums.dokumen <> ''", kec.ID).Count(&tercapaiK)

			var kelurahanDocsKadarkum []KelurahanDokumen
			for _, kel := range kec.Kelurahans {
				var docs []models.Kadarkum
				config.DB.Where("kelurahan_id = ?", kel.ID).Find(&docs)

				var dokList []string
				for _, d := range docs {
					if d.Dokumen != "" {
						dokList = append(dokList, d.Dokumen)
					}
				}

				tercapai := 0
				if len(dokList) > 0 {
					tercapai = 1
				}

				kelurahanDocsKadarkum = append(kelurahanDocsKadarkum, KelurahanDokumen{
					NamaKelurahan: kel.Name,
					AdaDokumen:    tercapai == 1,
					Dokumen:       dokList,
					Total:         1,
					Tercapai:      tercapai,
					Persentase:    hitungPersen(tercapai, 1),
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
			config.DB.Model(&models.PJA{}).
				Joins("JOIN kelurahans ON kelurahans.id = pjas.kelurahan_id").
				Where("kelurahans.kecamatan_id = ? AND pjas.dokumen <> ''", kec.ID).Count(&tercapaiP)

			var kelurahanDocsPja []KelurahanDokumen
			for _, kel := range kec.Kelurahans {
				var docs []models.PJA
				config.DB.Where("kelurahan_id = ?", kel.ID).Find(&docs)

				var dokList []string
				for _, d := range docs {
					if d.Dokumen != "" {
						dokList = append(dokList, d.Dokumen)
					}
				}

				tercapai := 0
				if len(dokList) > 0 {
					tercapai = 1
				}

				kelurahanDocsPja = append(kelurahanDocsPja, KelurahanDokumen{
					NamaKelurahan: kel.Name,
					AdaDokumen:    tercapai == 1,
					Dokumen:       dokList,
					Total:         1,
					Tercapai:      tercapai,
					Persentase:    hitungPersen(tercapai, 1),
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
				var paralegals []models.Paralegal
				// Query untuk mengambil semua paralegal yang terkait dengan kelurahan ini
				config.DB.Model(&models.Paralegal{}).
					Joins("JOIN posbankums ON posbankums.id = paralegals.posbankum_id").
					Where("posbankums.kelurahan_id = ?", kel.ID).
					Find(&paralegals)

				var paralegalData []ParalegalData
				for _, p := range paralegals {
					paralegalData = append(paralegalData, ParalegalData{
						Nama:    p.Nama,
						Dokumen: p.Dokumen,
					})
				}

				totalParalegalKec += len(paralegals)

				kelurahanParalegals = append(kelurahanParalegals, KelurahanParalegal{
					NamaKelurahan: kel.Name,
					Total:         len(paralegals),
					Paralegals:    paralegalData,
				})
			}

			paralegalKec = append(paralegalKec, KecamatanParalegal{
				NamaKecamatan: kec.Name,
				Total:         totalParalegalKec,
				Kelurahans:    kelurahanParalegals,
			})
			totalParalegalKab += totalParalegalKec
		}

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
	}

	c.HTML(http.StatusOK, "user_dashboard.html", data)
}
