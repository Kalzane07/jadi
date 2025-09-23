package controllers // ... (import statements)
import (
	// ...
	"net/http"      // <--- Pastikan ini ada
	"os"            // <--- Pastikan ini ada
	"path/filepath" // <--- Tambahkan ini

	"github.com/gin-gonic/gin"
)

// ... (fungsi controller lainnya)

// ================== SERVE DOKUMEN AMAN ==================
func ServeDokumen(c *gin.Context) {
	// 1. Ambil parameter dari URL
	tipe := c.Param("tipe")
	filename := c.Param("filename")

	// 2. Keamanan: Whitelist tipe folder yang diizinkan
	// Ini untuk mencegah orang mencoba mengakses folder lain seperti ../../
	allowedTypes := map[string]bool{
		"posbankum": true,
		"paralegal": true,
		"pja":       true,
		"kadarkum":  true,
	}
	if !allowedTypes[tipe] {
		c.String(http.StatusForbidden, "Tipe dokumen tidak valid")
		return
	}

	// 3. Keamanan: Bersihkan nama file dari karakter aneh (mencegah path traversal)
	// filepath.Base akan menghapus semua prefix direktori seperti ../
	cleanFilename := filepath.Base(filename)

	// 4. Gabungkan path menjadi path file yang sebenarnya di server
	fileLocation := filepath.Join("uploads", tipe, cleanFilename)

	// 5. Cek apakah file benar-benar ada
	if _, err := os.Stat(fileLocation); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "File tidak ditemukan")
		return
	}

	// 6. Sajikan file ke pengguna
	// c.File() secara otomatis mengatur Content-Type dan header lain yang diperlukan
	c.File(fileLocation)
}
