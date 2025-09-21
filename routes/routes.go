package routes

import (
	"go-admin/controllers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes untuk semua routing aplikasi
func SetupRoutes(r *gin.Engine) {
	// ================= LANDING PAGE & STATISTIK (NEW) =================
	r.GET("/", controllers.LandingPage)
	//	r.GET("/statistik-kabupaten", controllers.StatistikPageByKabupaten) // Rute baru untuk statistik per kabupaten

	// ================= AUTH =================
	r.GET("/login", controllers.ShowLogin)
	r.POST("/login", controllers.DoLogin)
	r.GET("/logout", controllers.Logout)

	// ================= ROUTES ADMIN =================
	admin := r.Group("/admin")
	admin.Use(controllers.AuthRequired(), controllers.RoleRequired("admin"))
	{
		// ================= DASHBOARD =================
		admin.GET("/", controllers.AdminPanel)

		// ================= POSBANKUM CRUD =================
		admin.GET("/posbankum", controllers.PosbankumIndex)
		admin.GET("/posbankum/create", controllers.PosbankumCreate)
		admin.POST("/posbankum/store", controllers.PosbankumStore)
		admin.GET("/posbankum/edit/:id", controllers.PosbankumEdit)
		admin.POST("/posbankum/update/:id", controllers.PosbankumUpdate)
		admin.GET("/posbankum/delete/:id", controllers.PosbankumDelete)

		// ================= PARALEGAL CRUD =================
		admin.GET("/paralegal", controllers.ParalegalIndex)
		admin.GET("/paralegal/create", controllers.ParalegalCreate)
		admin.POST("/paralegal/store", controllers.ParalegalStore)
		admin.GET("/paralegal/edit/:id", controllers.ParalegalEdit)
		admin.POST("/paralegal/update/:id", controllers.ParalegalUpdate)
		admin.GET("/paralegal/delete/:id", controllers.ParalegalDelete)

		// ================= KADARKUM CRUD =================
		admin.GET("/kadarkum", controllers.KadarkumIndex)
		admin.GET("/kadarkum/create", controllers.KadarkumCreate)
		admin.POST("/kadarkum/store", controllers.KadarkumStore)
		admin.GET("/kadarkum/edit/:id", controllers.KadarkumEdit)
		admin.POST("/kadarkum/update/:id", controllers.KadarkumUpdate)
		admin.GET("/kadarkum/delete/:id", controllers.KadarkumDelete)

		// ================= PJA CRUD =================
		admin.GET("/pja", controllers.PJAIndex)
		admin.GET("/pja/create", controllers.PJACreate)
		admin.POST("/pja/store", controllers.PJAStore)
		admin.GET("/pja/edit/:id", controllers.PJAEdit)
		admin.POST("/pja/update/:id", controllers.PJAUpdate)
		admin.GET("/pja/delete/:id", controllers.PJADelete)

		// ================= MASTER WILAYAH =================
		admin.GET("/provinsi", controllers.ProvinsiIndex)
		admin.GET("/kabupaten", controllers.KabupatenIndex)
		admin.GET("/kecamatan", controllers.KecamatanIndex)
		admin.GET("/kelurahan", controllers.KelurahanIndex)

		// ================= API (untuk autocomplete) =================
		admin.GET("/api/kelurahan/search", controllers.KelurahanSearch)
		admin.GET("/api/posbankum/search", controllers.PosbankumSearch) // 🔹 baru ditambah
	}

	// ================= ROUTES USER =================
	user := r.Group("/user")
	user.Use(controllers.AuthRequired(), controllers.RoleRequired("user"))
	{
		// 🔹 Semua data ditampilkan dalam 1 halaman
		user.GET("/", controllers.UserDashboard)
	}
}
