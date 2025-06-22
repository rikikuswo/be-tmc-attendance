package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// Ganti dengan konfigurasi database kamu
	dsn := "host=127.0.0.1 user=postgres password=admin12345 dbname=postgres port=5432 sslmode=disable"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Gagal koneksi ke database: ", err)
		os.Exit(1)
	}

	fmt.Println("Koneksi ke database berhasil!")

	DB = database
}
