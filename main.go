package main

import (
	"encoding/json"
	"fmt"
	"go-admin/config"
	"go-admin/controllers"
	"go-admin/routes"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// ============ HELPER FUNCTIONS LAMA ANDA (TETAP ADA) ============

func add(x, y int) int { return x + y }
func sub(x, y int) int { return x - y }
func iter(count int) []int {
	s := make([]int, count)
	for i := 0; i < count; i++ {
		s[i] = i
	}
	return s
}
func now() time.Time { return time.Now() }
func calcPersen(tercapai, total int) float64 {
	if total == 0 {
		return 0
	}
	return (float64(tercapai) / float64(total)) * 100
}
func calcTotal(a, b int) string {
	if b == 0 {
		return "0 / 0"
	}
	return fmt.Sprintf("%d / %d", a, b)
}
func totalTercapai(data []controllers.KecamatanSummary) int {
	total := 0
	for _, v := range data {
		total += v.Tercapai
	}
	return total
}
func totalKeseluruhan(data []controllers.KecamatanSummary) int {
	total := 0
	for _, v := range data {
		total += v.Total
	}
	return total
}
func isSlice(v interface{}) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
func toJSON(v interface{}) template.JS {
	a, err := json.Marshal(v)
	if err != nil {
		return template.JS("null")
	}
	return template.JS(a)
}

// ============ HELPER FUNCTIONS BARU (UNTUK DOKUMEN AMAN) ============

// ✅ Helper function ini sekarang menerima ID, bukan path
func formatDokumenURL(id uint) string {
	return fmt.Sprintf("/jadi/dokumen/view/%d", id) // ✅ Ubah path untuk menghindari konflik
}

// ✅ Helper function ini sekarang menerima data dokumen dan me-render link
func renderDokumenLink(dokumenData interface{}) template.HTML {
	if dokumenData == nil {
		return ""
	}
	// Periksa apakah dokumenData adalah objek model Posbankum, PJA, dll.
	val := reflect.ValueOf(dokumenData)
	if val.Kind() != reflect.Struct {
		return ""
	}

	idField := val.FieldByName("ID")
	if !idField.IsValid() {
		return ""
	}
	id, ok := idField.Interface().(uint)
	if !ok {
		return ""
	}

	url := formatDokumenURL(id)
	linkHTML := fmt.Sprintf(
		`<a href="%s" target="_blank" class="inline-flex items-center gap-1 text-xs bg-green-500 text-white px-2.5 py-1 rounded-full hover:bg-green-600 transition-colors">
            <i class="fas fa-file-alt"></i> Dokumen
        </a>`, url)
	return template.HTML(linkHTML)
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// ✅ Menambahkan logger middleware Gin untuk menampilkan detail setiap permintaan.
	r.Use(gin.Logger())

	if err := r.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		log.Fatal("Gagal set trusted proxies:", err)
	}

	funcMap := template.FuncMap{
		"add":               add,
		"sub":               sub,
		"iter":              iter,
		"now":               now,
		"calcPersen":        calcPersen,
		"calcTotal":         calcTotal,
		"totalTercapai":     totalTercapai,
		"totalKeseluruhan":  totalKeseluruhan,
		"isSlice":           isSlice,
		"toJSON":            toJSON,
		"hasPrefix":         strings.HasPrefix,
		"hasSuffix":         strings.HasSuffix,
		"formatDokumenURL":  formatDokumenURL,
		"renderDokumenLink": renderDokumenLink,
	}
	r.SetFuncMap(funcMap)
	r.LoadHTMLGlob("templates/*")

	// Session setup
	sessionKey := "kunci-rahasia-anda"
	store := cookie.NewStore([]byte(sessionKey))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 8,
		HttpOnly: true,
		Secure:   gin.Mode() == gin.ReleaseMode,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("mysession", store))

	config.ConnectDB()
	jadi := r.Group("/jadi")
	{
		routes.SetupRoutes(jadi)
		jadi.Static("/static", "./static")

		// ✅ HAPUS baris ini karena sudah ditangani di routes.go
		// jadi.GET("/dokumen/:id", controllers.ServeDocumentFromDB)
	}

	fmt.Println("✅ Server berjalan di http://localhost:8080/jadi")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Gagal menjalankan server:", err)
	}
}
