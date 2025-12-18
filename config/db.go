package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB global supaya bisa dipakai di controller
var DB *gorm.DB

func ConnectDB() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Gagal koneksi database:", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Gagal mendapatkan data dari GORM: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	// assign ke global
	DB = database
	// AutoMigrate semua model
	// err = DB.AutoMigrate(
	// 	&models.Provinsi{},
	// 	&models.Kabupaten{},
	// 	&models.Kecamatan{},
	// 	&models.Kelurahan{},
	// 	&models.Posbankum{},
	// 	&models.Paralegal{},
	// 	&models.PJA{},
	// 	&models.Kadarkum{},
	// 	&models.User{},
	// )
}
