package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"reflect"
	"strings"
	"time"

	"go-admin/config"
	"go-admin/controllers"
	"go-admin/routes"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// ============ helper functions untuk template ============

// tambah
func add(x, y int) int {
	return x + y
}

// kurang
func sub(x, y int) int {
	return x - y
}

// bikin slice [0..count-1]
func iter(count int) []int {
	s := make([]int, count)
	for i := 0; i < count; i++ {
		s[i] = i
	}
	return s
}

// helper untuk return waktu sekarang (buat {{ now.Year }} di template)
func now() time.Time {
	return time.Now()
}

// helper untuk hitung persentase (Tercapai / Total)
func calcPersen(tercapai, total int) float64 {
	if total == 0 {
		return 0
	}
	return (float64(tercapai) / float64(total)) * 100
}

// helper untuk format total string "a / b"
func calcTotal(a, b int) string {
	if b == 0 {
		return "0 / 0"
	}
	return fmt.Sprintf("%d / %d", a, b)
}

// helper hitung total Tercapai dari slice controllers.KecamatanSummary
func totalTercapai(data []controllers.KecamatanSummary) int {
	total := 0
	for _, v := range data {
		total += v.Tercapai
	}
	return total
}

// helper hitung total keseluruhan dari slice controllers.KecamatanSummary
func totalKeseluruhan(data []controllers.KecamatanSummary) int {
	total := 0
	for _, v := range data {
		total += v.Total
	}
	return total
}

// helper cek apakah value slice
func isSlice(v interface{}) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

// helper konversi ke JSON (buat {{ .Var | toJSON }} di template)
func toJSON(v interface{}) template.JS {
	a, err := json.Marshal(v)
	if err != nil {
		return template.JS("null")
	}
	return template.JS(a)
}

func main() {
	// init Gin
	r := gin.Default()

	// load HTML templates + register functions
	funcMap := template.FuncMap{
		"add":              add,
		"sub":              sub,
		"iter":             iter,
		"now":              now,
		"calcPersen":       calcPersen,
		"calcTotal":        calcTotal,
		"totalTercapai":    totalTercapai,
		"totalKeseluruhan": totalKeseluruhan,
		"isSlice":          isSlice,
		"hasPrefix":        strings.HasPrefix,
		"hasSuffix":        strings.HasSuffix,
		"toJSON":           toJSON, // ✅ tambahkan ini
	}
	r.SetFuncMap(funcMap)
	r.LoadHTMLGlob("templates/*")

	// session setup
	store := cookie.NewStore([]byte("super-secret-key")) // ganti dengan key yang lebih aman
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 8, // 8 jam
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("mysession", store))

	// ============ CONNECT DATABASE ============
	config.ConnectDB()

	// ============ SETUP ROUTES ============
	routes.SetupRoutes(r)

	// serve static files (logo, background, css/js custom)
	r.Static("/static", "./static")

	// serve file uploads (PDF, gambar, dll.)
	r.Static("/uploads", "./uploads")

	// run server di port 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatal("❌ Gagal menjalankan server:", err)
	}
}
