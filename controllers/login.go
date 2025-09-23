package controllers

import (
	"net/http"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ====== Utility ======
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// ====== Controller ======

// Login Page
func ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title": "Login",
	})
}

// Proses Login
func DoLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var user models.User
	// cari user berdasarkan username
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Title": "Login",
			"error": "❌ Username atau password salah", // Ubah pesan error untuk mencegah enumerasi pengguna
		})
		return
	}

	// cek password pakai bcrypt. Fallback plaintext dihapus!
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Title": "Login",
			"error": "❌ Username atau password salah", // Pesan error yang sama
		})
		return
	}

	// simpan session
	session := sessions.Default(c)
	session.Set("user", user.Username)
	session.Set("role", user.Role)
	session.Save()

	// redirect sesuai role
	switch user.Role {
	case "admin":
		c.Redirect(http.StatusFound, "/jadi/admin")
	case "user":
		c.Redirect(http.StatusFound, "/jadi/user")
	default:
		c.Redirect(http.StatusFound, "/jadi/login")
	}
}

// Logout
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/jadi/login")
}

// Middleware cek login (apapun role-nya)
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		if user == nil {
			c.Redirect(http.StatusFound, "/jadi/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

// Middleware cek role tertentu
func RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		role := session.Get("role")

		if role == nil {
			c.Redirect(http.StatusFound, "/jadi/login")
			c.Abort()
			return
		}

		allowed := false
		for _, r := range roles {
			if role == r {
				allowed = true
				break
			}
		}

		if !allowed {
			c.String(http.StatusForbidden, "🚫 Akses ditolak. Role Anda tidak punya izin.")
			c.Abort()
			return
		}

		c.Next()
	}
}
