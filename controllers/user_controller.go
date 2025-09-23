package controllers

import (
	"math"
	"net/http"
	"strconv"

	"go-admin/config"
	"go-admin/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ================= Util =================

// hashPassword -> bikin hash dari password plaintext
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
		"user":       "admin", // nanti bisa diganti sesuai session login
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

	// hash password
	hashed, err := hashPassword(password)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal hash password")
		return
	}

	user := models.User{
		Username: username,
		Password: hashed,
		Role:     role,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal simpan user")
		return
	}

	c.Redirect(http.StatusFound, "/jadi/admin/users")
}

// Show form edit
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

// Update user
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

	c.Redirect(http.StatusFound, "/jadi/admin/users")
}

// Hapus user
func UserDelete(c *gin.Context) {
	id := c.Param("id")
	idInt, _ := strconv.Atoi(id)

	if err := config.DB.Delete(&models.User{}, idInt).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal hapus user")
		return
	}

	c.Redirect(http.StatusFound, "/jadi/admin/users")
}
