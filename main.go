package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"be-tmc-attendance/database"
	"be-tmc-attendance/models"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/speps/go-hashids"
)

func main() {
	database.ConnectDB()

	// Auto migrate
	database.DB.AutoMigrate(&models.TMCAttendance{})

	r := mux.NewRouter()

	r.HandleFunc("/api/form", submitForm).Methods("POST")
	r.HandleFunc("/api/form/{id}", getFormByID).Methods("GET")
	r.HandleFunc("/api/attendees", getAttendees).Methods("GET")
	r.HandleFunc("/api/attend/{id}", attendHandler).Methods("POST")
	r.HandleFunc("/api/all-attendees", allAttendeesHandler).Methods("GET")


	handler := cors.Default().Handler(r)

	fmt.Println("Server running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", handler))
}

func encryptID(id uint) string {
	hd := hashids.NewData()
	hd.Salt = "secret_salt"
	hd.MinLength = 8
	h, _ := hashids.NewWithData(hd)

	hash, _ := h.Encode([]int{int(id)})
	return hash
}

// Handler untuk menerima form
func submitForm(w http.ResponseWriter, r *http.Request) {
	var form models.TMCAttendance

	err := json.NewDecoder(r.Body).Decode(&form)
	if err != nil {
		http.Error(w, "Input tidak valid", http.StatusBadRequest)
		return
	}

	// ✅ Validasi Field
	if len(form.Name) < 3 {
		http.Error(w, "Nama minimal 3 karakter", http.StatusBadRequest)
		return
	}

	if len(form.Company) < 3 {
		http.Error(w, "Nama perusahaan minimal 3 karakter", http.StatusBadRequest)
		return
	}

	// allowedStatus := map[string]bool{
	// 	"Observer":    true,
	// 	"Participant": true,
	// 	"Speaker":     true,
	// }

	// if _, ok := allowedStatus[form.Status]; !ok {
	// 	http.Error(w, "Status tidak valid", http.StatusBadRequest)
	// 	return
	// }

	// ✅ Cek Data Duplikat
	var existing models.TMCAttendance
	result := database.DB.Where("name = ? AND company = ? AND status = ?", form.Name, form.Company, form.Status).First(&existing)

	if result.RowsAffected > 0 {
		http.Error(w, "Data sudah pernah diinput.", http.StatusBadRequest)
		return
	}

	// ✅ Simpan Data Jika Belum Ada
	createResult := database.DB.Create(&form)
	if createResult.Error != nil {
		http.Error(w, "Gagal menyimpan data", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Data disimpan: %+v\n", form)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Data berhasil disimpan!",
		"id": encryptID(form.ID),
	})
}


func getFormByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	encryptedID := params["id"]

	// Decode ID
	hd := hashids.NewData()
	hd.Salt = "secret_salt"
	hd.MinLength = 8

	h, _ := hashids.NewWithData(hd)
	decoded, err := h.DecodeWithError(encryptedID)
	if err != nil || len(decoded) == 0 {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	realID := decoded[0]

	// Cari data berdasarkan ID asli
	var form models.TMCAttendance
	result := database.DB.First(&form, realID)

	if result.Error != nil {
		http.Error(w, "Data tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(form)
}

func getAttendees(w http.ResponseWriter, r *http.Request) {
	var attendees []models.TMCAttendance
	database.DB.Where("attended_at IS NOT NULL").Order("attended_at DESC").Find(&attendees)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attendees)
}

func attendHandler(w http.ResponseWriter, r *http.Request) {
    // Ambil ID dari parameter URL
    params := mux.Vars(r)
    id := params["id"]

    var attendance models.TMCAttendance

    // Cari data peserta berdasarkan ID
    result := database.DB.First(&attendance, id)
    if result.Error != nil {
        http.Error(w, "Data tidak ditemukan", http.StatusNotFound)
        return
    }

    // Validasi apakah peserta sudah hadir
    if attendance.AttendedAt.Valid {
        http.Error(w, "Peserta sudah hadir", http.StatusBadRequest)
        return
    }

    // Update attended_at dengan waktu sekarang
    attendance.AttendedAt = sql.NullTime{
        Time:  time.Now(),
        Valid: true,
    }

    // Simpan perubahan
    saveResult := database.DB.Save(&attendance)
    if saveResult.Error != nil {
        http.Error(w, "Gagal mencatat kehadiran", http.StatusInternalServerError)
        return
    }

    // Berikan respon sukses
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Kehadiran berhasil dicatat",
        "data":    attendance,
    })
}

func allAttendeesHandler(w http.ResponseWriter, r *http.Request) {
    var attendees []models.TMCAttendance

    // Ambil semua data peserta
    result := database.DB.Order("created_at desc").Find(&attendees)
    if result.Error != nil {
        http.Error(w, "Gagal mengambil data peserta", http.StatusInternalServerError)
        return
    }

    // Berikan respon sukses
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message":   "Data peserta berhasil diambil",
        "attendees": attendees,
    })
}





