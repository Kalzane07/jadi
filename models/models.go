package models

import (
	"time"
)

// ================= Master Data =================

// Provinsi
type Provinsi struct {
	ID        uint   `gorm:"primaryKey"`
	Code      string `gorm:"unique;not null"`
	Name      string `gorm:"not null"`
	CreatedAt *time.Time
	UpdatedAt *time.Time

	Kabupatens []Kabupaten `gorm:"foreignKey:ProvinsiID"`
}

// Kabupaten
type Kabupaten struct {
	ID         uint   `gorm:"primaryKey"`
	Code       string `gorm:"unique;not null"`
	Name       string `gorm:"not null"`
	ProvinsiID uint   `gorm:"not null"`
	CreatedAt  *time.Time
	UpdatedAt  *time.Time

	Provinsi   Provinsi    // ✅ biar bisa Preload("Provinsi")
	Kecamatans []Kecamatan `gorm:"foreignKey:KabupatenID"`
}

// Kecamatan
type Kecamatan struct {
	ID          uint   `gorm:"primaryKey"`
	Code        string `gorm:"unique;not null"`
	Name        string `gorm:"not null"`
	KabupatenID uint   `gorm:"not null"`
	CreatedAt   *time.Time
	UpdatedAt   *time.Time

	Kabupaten  Kabupaten   // ✅ biar bisa Preload("Kabupaten")
	Kelurahans []Kelurahan `gorm:"foreignKey:KecamatanID"`
}

// Kelurahan
type Kelurahan struct {
	ID          uint   `gorm:"primaryKey"`
	Code        string `gorm:"unique;not null"`
	Name        string `gorm:"not null"`
	KecamatanID uint   `gorm:"not null"`
	CreatedAt   *time.Time
	UpdatedAt   *time.Time

	Kecamatan  Kecamatan   // ✅ biar bisa Preload("Kecamatan")
	Posbankums []Posbankum `gorm:"foreignKey:KelurahanID"`
	Kadarkums  []Kadarkum  `gorm:"foreignKey:KelurahanID"`
	PJAs       []PJA       `gorm:"foreignKey:KelurahanID"`
}

// ================= Entity Utama =================

// Posbankum
type Posbankum struct {
	ID          uint   `gorm:"primaryKey"`
	KelurahanID uint   `gorm:"not null"`
	Dokumen     string `gorm:"type:text;not null"`
	Catatan     string `gorm:"type:text"`
	CreatedAt   *time.Time
	UpdatedAt   *time.Time

	Kelurahan  Kelurahan
	Paralegals []Paralegal `gorm:"foreignKey:PosbankumID"`
}

// Paralegal
type Paralegal struct {
	ID          uint   `gorm:"primaryKey"`
	PosbankumID uint   `gorm:"not null"`
	Nama        string `gorm:"not null"`
	Dokumen     string `gorm:"type:text"`
	CreatedAt   *time.Time
	UpdatedAt   *time.Time

	Posbankum Posbankum
}

// PJA
type PJA struct {
	ID          uint   `gorm:"primaryKey"`
	KelurahanID uint   `gorm:"not null"`
	Dokumen     string `gorm:"type:text;not null"`
	Catatan     string `gorm:"type:text"`
	CreatedAt   *time.Time
	UpdatedAt   *time.Time

	Kelurahan Kelurahan
}

// Kadarkum
type Kadarkum struct {
	ID          uint   `gorm:"primaryKey"`
	KelurahanID uint   `gorm:"not null"`
	Dokumen     string `gorm:"type:text;not null"`
	Catatan     string `gorm:"type:text"`
	CreatedAt   *time.Time
	UpdatedAt   *time.Time

	Kelurahan Kelurahan
}

// ================= Auth =================

// User
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	Role      string `gorm:"type:enum('admin','user');default:'user'"`
	CreatedAt *time.Time
}
