package controllers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"regexp"

	"golang.org/x/crypto/bcrypt"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-gonic/gin"
)

// ================= Util =================

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// ================= CRUD =================

// Index -> list semua user dengan pagination + search
func UserIndex(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit := 10
	offset := (page - 1) * limit
	search := c.Query("q")

	var users []models.User
	db := config.DB.Model(&models.User{})

	if search != "" {
		like := "%" + search + "%"
		db = db.Where("username LIKE ?", like)
	}

	var total int64
	db.Count(&total)

	db.Offset(offset).Limit(limit).Find(&users)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.HTML(http.StatusOK, "user_index.html", gin.H{
		"Title":      "Manajemen User",
		"Users":      users,
		"Search":     search,
		"Page":       page,
		"TotalPages": totalPages,
		"Offset":     offset,
		"user":       c.GetString("user"),
	})
}

// Show form tambah user
func UserCreateForm(c *gin.Context) {
	c.HTML(http.StatusOK, "user_create.html", gin.H{
		"Title": "Tambah User",
	})
}

// Simpan user baru
func UserCreate(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	role := c.PostForm("role")

	// Validasi password menggunakan fungsi validatePassword
	if err := validatePassword(password); err != nil {
		log.Printf("Validasi password gagal: %v", err)
		c.HTML(http.StatusBadRequest, "user_create.html", gin.H{

			"Title":         "Tambah User",
			"ErrorPassword": err.Error(),
		})
		return
	}

	// Hash password menggunakan bcrypt
	hashed, err := hashPassword(password)
	if err != nil {
		log.Printf("Gagal hash password: %v", err)
		c.String(http.StatusInternalServerError, "Gagal hash password")
		return
	}

	// Membuat objek user baru
	user := models.User{
		Username: username,
		Password: hashed,
		Role:     role,
	}

	// Cek apakah username sudah ada
	var existingUser models.User
	if err := config.DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
		log.Printf("Username sudah ada: %s", username)
		c.HTML(http.StatusBadRequest, "user_create.html", gin.H{
			"Title":         "Tambah User",
			"ErrorUsername": "Username sudah ada",

			"Username": username,
		})
		return
	}

	// Menyimpan user ke database
	if err := config.DB.Create(&user).Error; err != nil {
		log.Printf("Gagal simpan user ke database: %v", err)
		c.String(http.StatusInternalServerError, "Gagal simpan user")
		return
	}
	c.Redirect(http.StatusFound, "/admin/users")
}

// validatePassword -> validasi password
func validatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password harus minimal 6 karakter")
	}
	if !hasUpperCase(password) || !hasSymbol(password) {
		return fmt.Errorf("password harus mengandung setidaknya 1 huruf besar dan 1 simbol")
	}
	return nil

}
func hasUpperCase(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}

	return false
}

// hasSymbol -> cek apakah string mengandung simbol
func hasSymbol(s string) bool {
	// Definisi regular expression untuk simbol yang diizinkan
	pattern := `[!@#$%^&*()]`

	// Mencocokkan string dengan regular expression
	matched, _ := regexp.MatchString(pattern, s)
	return matched

}

// UserEditForm -> menampilkan form edit user
func UserEditForm(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.String(http.StatusNotFound, "User tidak ditemukan")
		return
	}

	// kirim data user dengan ID ke template
	c.HTML(http.StatusOK, "user_edit.html", gin.H{
		"Title": "Edit User",
		"User":  user,
	})
}

// UserUpdate -> update data user
func UserUpdate(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.String(http.StatusNotFound, "User tidak ditemukan")
		return
	}

	// Username tidak bisa diubah → abaikan input username
	password := c.PostForm("password")
	role := c.PostForm("role")

	user.Role = role

	// update password kalau diisi
	if password != "" {
		hashed, err := hashPassword(password)
		if err != nil {
			c.String(http.StatusInternalServerError, "Gagal hash password")
			return
		}
		user.Password = hashed
	}

	config.DB.Save(&user)

	c.Redirect(http.StatusFound, "/admin/users")
}

// UserDelete -> hapus user
func UserDelete(c *gin.Context) {
	id := c.Param("id")
	idInt, _ := strconv.Atoi(id)

	if err := config.DB.Delete(&models.User{}, idInt).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal hapus user")
		return
	}

	c.Redirect(http.StatusFound, "/admin/users")
}
