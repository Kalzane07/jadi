package routes

import (
	"go-admin/controllers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes untuk semua routing aplikasi
func SetupRoutes(jadi *gin.RouterGroup) {
	// ================= LANDING PAGE & STATISTIK =================
	jadi.GET("/", controllers.LandingPage)

	// ================= AUTH =================
	jadi.GET("/login", controllers.ShowLogin)
	jadi.POST("/login", controllers.DoLogin)
	jadi.GET("/logout", controllers.Logout)

	// ================= ROUTES ADMIN (UNTUK HALAMAN WEB) =================
	// Grup ini khusus untuk halaman-halaman yang merender HTML dan butuh role "admin".
	admin := jadi.Group("/admin")
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
	}

	// ================= ROUTES API (UNTUK DATA JSON) =================
	// Grup ini khusus untuk endpoint API yang mengembalikan data JSON.
	// Cukup pakai AuthRequired() saja, karena tidak merender halaman admin.
	api := jadi.Group("/api")
	api.Use(controllers.AuthRequired())
	{
		api.GET("/kelurahan/search", controllers.KelurahanSearch)
		api.GET("/posbankum/search", controllers.PosbankumSearch)
		api.GET("/kadarkum/search", controllers.PosbankumSearch)
		api.GET("/pja/search", controllers.PosbankumSearch)
		api.GET("/paralegal/search", controllers.PosbankumSearch)
	}

	// ================= ROUTES USER =================
	user := jadi.Group("/user")
	user.Use(controllers.AuthRequired(), controllers.RoleRequired("user"))
	{
		user.GET("/", controllers.UserDashboard)
	}
}
