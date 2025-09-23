package config

import (
	"fmt"
	"go-admin/models"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB global supaya bisa dipakai di controller
var DB *gorm.DB

func ConnectDB() {
	// sesuaikan user:password@tcp(host:port)/dbname
	dsn := "root:@tcp(127.0.0.1:3306)/admingo?charset=utf8mb4&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Gagal koneksi database:", err)
	}

	// assign ke global
	DB = database

	// AutoMigrate semua model
	err = DB.AutoMigrate(
		&models.Provinsi{},
		&models.Kabupaten{},
		&models.Kecamatan{},
		&models.Kelurahan{},
		&models.Posbankum{},
		&models.Paralegal{},
		&models.PJA{},
		&models.Kadarkum{},
		&models.User{},
	)
	if err != nil {
		log.Fatal("❌ AutoMigrate error:", err)
	}

	fmt.Println("✅ Database connected & migrated!")
}
