package main

import (
	"encoding/json"
	"fmt"
	"go-admin/config"
	"go-admin/controllers"
	"go-admin/routes"
	"html/template"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// ============ helper functions untuk template ============
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
func isSlice(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
func toJSON(v any) template.JS {
	a, err := json.Marshal(v)
	if err != nil {
		return template.JS("null")
	}
	return template.JS(a)
}

func main() {
	// ============ SET GIN MODE KE RELEASE ============
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Logger())

	if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		log.Fatal("Gagal set trusted proxies:", err)
	}

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
		"toJSON":           toJSON,
	}
	r.SetFuncMap(funcMap)
	r.LoadHTMLGlob("templates/*")

	// session setup
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 8,
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("mysession", store))

	// ============ CONNECT DATABASE ============
	config.ConnectDB()

	// ============ SETUP ROUTES ============
	jadi := r.Group("/jadi")
	{
		// route app
		routes.SetupRoutes(jadi)

		// serve static files & uploads di bawah /jadi
		jadi.Static("/static", "./static")
		jadi.Static("/uploads", "./uploads")
	}
	if err := r.Run(":8182"); err != nil {
		log.Fatal("Gagal menjalankan server:", err)
	}
}
