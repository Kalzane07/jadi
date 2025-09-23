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

func formatDokumenURL(path string) string {
	if path == "" {
		return "#"
	}
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		log.Printf("Format path dokumen tidak valid: %s", path)
		return "#"
	}
	return fmt.Sprintf("/jadi/dokumen/%s/%s", parts[1], parts[2])
}

func renderDokumenLinks(dokumenData interface{}) template.HTML {
	if dokumenData == nil {
		return ""
	}
	if reflect.TypeOf(dokumenData).Kind() == reflect.Slice {
		slice := reflect.ValueOf(dokumenData)
		var links strings.Builder
		for i := 0; i < slice.Len(); i++ {
			path, ok := slice.Index(i).Interface().(string)
			if !ok {
				continue
			}
			url := formatDokumenURL(path)
			linkHTML := fmt.Sprintf(
				`<a href="%s" target="_blank" class="inline-flex items-center gap-1 text-xs bg-green-500 text-white px-2.5 py-1 rounded-full hover:bg-green-600 transition-colors mr-1 mb-1">
					<i class="fas fa-file-alt"></i> Dokumen
				</a>`, url)
			links.WriteString(linkHTML)
		}
		return template.HTML(links.String())
	}
	if path, ok := dokumenData.(string); ok && path != "" {
		url := formatDokumenURL(path)
		linkHTML := fmt.Sprintf(
			`<a href="%s" target="_blank" class="inline-flex items-center gap-1 text-xs bg-green-500 text-white px-2.5 py-1 rounded-full hover:bg-green-600 transition-colors">
				<i class="fas fa-file-alt"></i> Dokumen
			</a>`, url)
		return template.HTML(linkHTML)
	}
	return ""
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	if err := r.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		log.Fatal("Gagal set trusted proxies:", err)
	}

	funcMap := template.FuncMap{
		"add":                add,
		"sub":                sub,
		"iter":               iter,
		"now":                now,
		"calcPersen":         calcPersen,
		"calcTotal":          calcTotal,
		"totalTercapai":      totalTercapai,
		"totalKeseluruhan":   totalKeseluruhan,
		"isSlice":            isSlice,
		"toJSON":             toJSON,
		"hasPrefix":          strings.HasPrefix,
		"hasSuffix":          strings.HasSuffix,
		"formatDokumenURL":   formatDokumenURL,
		"renderDokumenLinks": renderDokumenLinks,
	}
	r.SetFuncMap(funcMap)
	r.LoadHTMLGlob("templates/*")

	// Session setup
	// Kunci sesi sekarang di hardcode untuk menghindari error variabel lingkungan.
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
	}

	fmt.Println("✅ Server berjalan di http://localhost:8080/jadi")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Gagal menjalankan server:", err)
	}
}
