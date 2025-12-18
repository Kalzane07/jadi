package controllers

import (
	"net/http"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-gonic/gin"
)

// File ini berisi handler untuk halaman publik yang tidak memerlukan autentikasi.

// Struct untuk data peta
type MapData struct {
	Nama      string  `json:"nama"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Posbankum int     `json:"posbankum"`
	Kadarkum  int     `json:"kadarkum"`
	PJA       int     `json:"pja"`
	Paralegal int     `json:"paralegal"`
}

// Struct untuk data testimoni
type Testimonial struct {
	Nama    string
	Jabatan string
	Kutipan string
	FotoURL string
}

// Struct untuk data timeline
type TimelineEvent struct {
	Year    string
	Title   string
	Content string
	Icon    string // Font Awesome icon class
}

// ==================== STRUCT BARU UNTUK PARALEGAL ====================

// Struktur untuk menampung nama dan dokumen Paralegal
type PublicParalegalData struct {
	ID      uint // ID diperlukan untuk link dokumen di template
	Nama    string
	Dokumen string
}

// Struktur untuk data Paralegal per kelurahan
type PublicKelurahanParalegal struct {
	NamaKelurahan string
	Total         int
	Paralegals    []PublicParalegalData
}

// Struktur untuk data Paralegal per kecamatan
type PublicKecamatanParalegal struct {
	NamaKecamatan string
	Total         int
	Kelurahans    []PublicKelurahanParalegal
}

// Struktur untuk data Paralegal per kabupaten
type PublicKabupatenParalegal struct {
	NamaKabupaten string
	Total         int
	Kecamatans    []PublicKecamatanParalegal
}

// ==================== STRUCT LAINNYA ====================

// Struktur buat nampung data dokumen per kelurahan
type PublicKelurahanDokumen struct {
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
type PublicKecamatanSummary struct {
	NamaKecamatan string
	Total         int
	Tercapai      int
	Persentase    float64
	Kelurahans    []PublicKelurahanDokumen
}

// Struktur buat nampung data summary per kabupaten
type PublicKabupatenSummary struct {
	NamaKabupaten string
	Total         int
	Tercapai      int
	Persentase    float64
	Kecamatans    []PublicKecamatanSummary
}

// Struktur buat data dashboard user
type PublicDashboardData struct {
	Title                   string
	Provinsi                string
	Posbankum               []PublicKabupatenSummary
	Kadarkum                []PublicKabupatenSummary
	PJA                     []PublicKabupatenSummary
	Paralegal               []PublicKabupatenParalegal
	TotalPosbankumProvinsi  int
	TotalKadarkumProvinsi   int
	TotalPjaProvinsi        int
	TotalParalegalProvinsi  int
	TotalKelurahanProvinsi  int
	PersenPosbankumProvinsi float64
	PersenKadarkumProvinsi  float64
	PersenPjaProvinsi       float64
	// AllKabupatens, BaseHref, dan IsPublicView tidak lagi diperlukan
}

// ==================== CONTROLLER ====================

func LandingPage(c *gin.Context) {
	var totalPosbankum, totalKadarkum, totalPja, totalParalegal int64

	config.DB.Model(&models.Posbankum{}).Count(&totalPosbankum)
	config.DB.Model(&models.Kadarkum{}).Count(&totalKadarkum)
	config.DB.Model(&models.Pja{}).Count(&totalPja)
	config.DB.Model(&models.Paralegal{}).Count(&totalParalegal)

	// Data dummy untuk testimonials
	testimonials := []Testimonial{
		{
			Nama:    "Ahmad Suryadi",
			Jabatan: "Kepala Desa Makmur",
			Kutipan: "Sistem JADI sangat membantu kami dalam memetakan kebutuhan hukum di desa. Data yang disajikan akurat dan mudah diakses, membuat perencanaan program penyuluhan menjadi lebih efektif dan tepat sasaran.",
			FotoURL: "/static/testimonials/person1.jpg",
		},
		{
			Nama:    "Siti Nurhaliza",
			Jabatan: "Paralegal Desa Sejahtera",
			Kutipan: "Sebagai paralegal, saya merasa lebih terhubung dengan Kanwil. Pelaporan kegiatan menjadi lebih mudah dan transparan. Ini adalah langkah maju yang luar biasa untuk pembinaan hukum di Jambi.",
			FotoURL: "/static/testimonials/person2.jpg",
		},
		{
			Nama:    "Budi Santoso",
			Jabatan: "Penyuluh Hukum",
			Kutipan: "Dengan JADI, kami bisa melihat gambaran besar cakupan program kami. Analisis data menjadi lebih mudah, memungkinkan kami untuk fokus pada wilayah yang paling membutuhkan perhatian.",
			FotoURL: "/static/testimonials/person3.jpg",
		},
	}

	// Data dummy untuk timeline
	timelineEvents := []TimelineEvent{
		{Year: "2022", Title: "Awal Program Digitalisasi", Content: "Kanwil Kemenkumham Jambi memulai inisiatif untuk mendigitalisasi data pembinaan hukum.", Icon: "fa-rocket"},
		{Year: "2023", Title: "Peluncuran JADI 1.0", Content: "Versi pertama Jambi Database Informasi (JADI) diluncurkan untuk internal, memusatkan data Posbankum dan Kadarkum.", Icon: "fa-database"},
		{Year: "2024", Title: "Ekspansi Fitur", Content: "JADI diperluas dengan menambahkan modul untuk Peacemaker Justice Award (PJA) dan data Paralegal terdaftar.", Icon: "fa-users"},
		{Year: "2025", Title: "JADI 2.0: Go Public!", Content: "Halaman publik dengan peta interaktif dan dasbor statistik diluncurkan untuk transparansi dan akses informasi yang lebih luas.", Icon: "fa-globe-asia"},
	}

	c.HTML(http.StatusOK, "landing.html", gin.H{
		"TotalPosbankum": totalPosbankum,
		"TotalKadarkum":  totalKadarkum,
		"TotalPja":       totalPja,
		"TotalParalegal": totalParalegal,
		"Testimonials":   testimonials,
		"TimelineEvents": timelineEvents,
	})
}

func PublicDashboard(c *gin.Context) {
	var provinsi models.Provinsi

	if err := config.DB.Preload("Kabupatens.Kecamatans.Kelurahans").First(&provinsi).Error; err != nil {
		c.String(http.StatusInternalServerError, "❌ Tidak ada provinsi di database")
		return
	}

	var (
		posbankumAll, kadarkumAll, pjaAll []PublicKabupatenSummary
		paralegalAll                      []PublicKabupatenParalegal
		totalPosProv, tercapaiPosProv     int
		totalKadProv, tercapaiKadProv     int
		totalPJAProv, tercapaiPJAProv     int
		totalParalegalProv                int
		totalKelurahanProv64              int64
	)

	config.DB.Model(&models.Kelurahan{}).Count(&totalKelurahanProv64)
	totalKelurahanProv := int(totalKelurahanProv64)

	for _, kab := range provinsi.Kabupatens {
		var posbankumKec, kadarkumKec, pjaKec []PublicKecamatanSummary
		var paralegalKec []PublicKecamatanParalegal

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

			var kelurahanDocsPos []PublicKelurahanDokumen
			for _, kel := range kec.Kelurahans {
				var posbankums []models.Posbankum
				config.DB.Where("kelurahan_id = ?", kel.ID).Find(&posbankums)

				tercapai := 0
				if len(posbankums) > 0 {
					tercapai = 1
				}

				kelurahanDocsPos = append(kelurahanDocsPos, PublicKelurahanDokumen{
					NamaKelurahan: kel.Name,
					Total:         1,
					Tercapai:      tercapai,
					Persentase:    hitungPersen(tercapai, 1),
					Posbankums:    posbankums,
				})
			}

			posbankumKec = append(posbankumKec, PublicKecamatanSummary{
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

			var kelurahanDocsKadarkum []PublicKelurahanDokumen
			for _, kel := range kec.Kelurahans {
				var kadarkums []models.Kadarkum
				config.DB.Where("kelurahan_id = ?", kel.ID).Find(&kadarkums)
				tercapai := 0
				if len(kadarkums) > 0 {
					tercapai = 1
				}
				kelurahanDocsKadarkum = append(kelurahanDocsKadarkum, PublicKelurahanDokumen{
					NamaKelurahan: kel.Name,
					Total:         1,
					Tercapai:      tercapai,
					Persentase:    hitungPersen(tercapai, 1),
					Kadarkums:     kadarkums,
				})
			}

			kadarkumKec = append(kadarkumKec, PublicKecamatanSummary{
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

			var kelurahanDocsPja []PublicKelurahanDokumen
			for _, kel := range kec.Kelurahans {
				var pjas []models.Pja
				config.DB.Where("kelurahan_id = ?", kel.ID).Find(&pjas)
				tercapai := 0
				if len(pjas) > 0 {
					tercapai = 1
				}
				kelurahanDocsPja = append(kelurahanDocsPja, PublicKelurahanDokumen{
					NamaKelurahan: kel.Name,
					Total:         1,
					Tercapai:      tercapai,
					Persentase:    hitungPersen(tercapai, 1),
					Pjas:          pjas,
				})
			}

			pjaKec = append(pjaKec, PublicKecamatanSummary{
				NamaKecamatan: kec.Name,
				Total:         int(totalP),
				Tercapai:      int(tercapaiP),
				Persentase:    hitungPersen(int(tercapaiP), int(totalP)),
				Kelurahans:    kelurahanDocsPja,
			})
			totalPJAKab += int(totalP)
			tercapaiPJAKab += int(tercapaiP)

			// ================== PARALEGAL ==================
			var kelurahanParalegals []PublicKelurahanParalegal
			totalParalegalKec := 0

			for _, kel := range kec.Kelurahans {
				var paralegalsFromDB []models.Paralegal
				config.DB.Model(&models.Paralegal{}).
					Joins("JOIN posbankums ON posbankums.id = paralegals.posbankum_id").
					Where("posbankums.kelurahan_id = ?", kel.ID).
					Find(&paralegalsFromDB)

				var paralegalDataForKelurahan []PublicParalegalData
				for _, p := range paralegalsFromDB {
					paralegalDataForKelurahan = append(paralegalDataForKelurahan, PublicParalegalData{
						ID:      p.ID, // Kirim ID ke template
						Nama:    p.Nama,
						Dokumen: p.Dokumen,
					})
				}

				kelurahanParalegals = append(kelurahanParalegals, PublicKelurahanParalegal{
					NamaKelurahan: kel.Name,
					Total:         len(paralegalsFromDB),
					Paralegals:    paralegalDataForKelurahan,
				})
				totalParalegalKec += len(paralegalsFromDB)
			}

			paralegalKec = append(paralegalKec, PublicKecamatanParalegal{
				NamaKecamatan: kec.Name,
				Total:         totalParalegalKec,
				Kelurahans:    kelurahanParalegals,
			})
			totalParalegalKab += totalParalegalKec

		}

		// Push ke level kabupaten
		posbankumAll = append(posbankumAll, PublicKabupatenSummary{
			NamaKabupaten: kab.Name,
			Total:         totalPosbankumKab,
			Tercapai:      tercapaiPosbankumKab,
			Persentase:    hitungPersen(tercapaiPosbankumKab, totalPosbankumKab),
			Kecamatans:    posbankumKec,
		})

		kadarkumAll = append(kadarkumAll, PublicKabupatenSummary{
			NamaKabupaten: kab.Name,
			Total:         totalKadarkumKab,
			Tercapai:      tercapaiKadarkumKab,
			Persentase:    hitungPersen(tercapaiKadarkumKab, totalKadarkumKab),
			Kecamatans:    kadarkumKec,
		})

		pjaAll = append(pjaAll, PublicKabupatenSummary{
			NamaKabupaten: kab.Name,
			Total:         totalPJAKab,
			Tercapai:      tercapaiPJAKab,
			Persentase:    hitungPersen(tercapaiPJAKab, totalPJAKab),
			Kecamatans:    pjaKec,
		})

		paralegalAll = append(paralegalAll, PublicKabupatenParalegal{
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

	data := PublicDashboardData{
		Title:                   "Detail Data Pembinaan Hukum",
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
		// AllKabupatens dan IsPublicView dihapus
	}

	c.HTML(http.StatusOK, "public_detail.html", data)
}

// Struct baru untuk detail kecamatan di peta
type KecamatanMapDetail struct {
	NamaKecamatan           string `json:"nama_kecamatan"`
	TotalTercapaiKecamatan  int    `json:"total_tercapai_kecamatan"`
	TotalKelurahanKecamatan int    `json:"total_kelurahan_kecamatan"`
}

// Struct baru untuk data peta yang lebih detail
type MapDetailData struct {
	NamaKabupaten           string               `json:"nama_kabupaten"`
	Lat                     float64              `json:"lat"`
	Lon                     float64              `json:"lon"`
	TotalTercapaiKabupaten  int                  `json:"total_tercapai_kabupaten"`
	TotalKelurahanKabupaten int                  `json:"total_kelurahan_kabupaten"`
	Kecamatans              []KecamatanMapDetail `json:"kecamatans"`
}

// MapDataAPI menyediakan data untuk peta interaktif
func MapDataAPI(c *gin.Context) {
	var kabupatens []models.Kabupaten
	if err := config.DB.Preload("Kecamatans.Kelurahans").Find(&kabupatens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kabupaten"})
		return
	}

	// Koordinat manual untuk setiap kabupaten di Jambi
	coords := map[string][2]float64{
		"Kabupaten  Merangin":            {-2.062, 102.13},
		"Kabupaten  Muaro Jambi":         {-1.583, 103.85},
		"Kabupaten Batanghari":           {-1.716, 103.26},
		"Kabupaten Bungo":                {-1.52, 102.1},
		"Kabupaten Kerinci":              {-2.08, 101.48},
		"Kabupaten Sarolangun":           {-2.25, 102.63},
		"Kabupaten Tanjung Jabung Barat": {-1.0, 103.46},
		"Kabupaten Tanjung Jabung Timur": {-1.13, 104.35},
		"Kabupaten Tebo":                 {-1.4, 102.31},
		"Kota Jambi":                     {-1.59, 103.61},
		"Kota Sungai Penuh":              {-2.06, 101.39},
	}

	var mapData []MapDetailData

	for _, kab := range kabupatens {
		var kecamatanDetails []KecamatanMapDetail
		totalTercapaiKab := 0
		totalKelurahanKab := 0

		for _, kec := range kab.Kecamatans {
			var tercapaiKec int64
			totalKelurahanKec := len(kec.Kelurahans)

			config.DB.Model(&models.Posbankum{}).
				Joins("JOIN kelurahans ON kelurahans.id = posbankums.kelurahan_id").
				Where("kelurahans.kecamatan_id = ?", kec.ID).Count(&tercapaiKec)

			kecamatanDetails = append(kecamatanDetails, KecamatanMapDetail{
				NamaKecamatan:           kec.Name,
				TotalTercapaiKecamatan:  int(tercapaiKec),
				TotalKelurahanKecamatan: totalKelurahanKec,
			})
			totalTercapaiKab += int(tercapaiKec)
			totalKelurahanKab += totalKelurahanKec
		}

		mapData = append(mapData, MapDetailData{
			NamaKabupaten:           kab.Name,
			Lat:                     coords[kab.Name][0],
			Lon:                     coords[kab.Name][1],
			TotalTercapaiKabupaten:  totalTercapaiKab,
			TotalKelurahanKabupaten: totalKelurahanKab,
			Kecamatans:              kecamatanDetails,
		})
	}

	c.JSON(http.StatusOK, mapData)
}
