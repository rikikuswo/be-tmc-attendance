package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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
	database.DB.AutoMigrate(&models.TMCAttendance{}, &models.TMCStatus{}, &models.TMCCompany{})

	r := mux.NewRouter()

	r.HandleFunc("/api/get-companies", getCompanies).Methods("GET")
	r.HandleFunc("/api/get-statuses", getStatuses).Methods("GET")

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

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateUniqueNumericID() (string, error) {
	for {
		// Generate angka random 5 digit (range: 10000 - 99999)
		newID := fmt.Sprintf("%05d", rand.Intn(90000)+10000)

		var existing models.TMCAttendance
		result := database.DB.Where("unique_id = ?", newID).First(&existing)

		if result.RowsAffected == 0 {
			return newID, nil // ID unik ditemukan
		}

		// Optional: Logging jika terjadi duplikasi
		fmt.Println("Duplicate ID found, regenerating...")
	}
}

func getCompanies(w http.ResponseWriter, r *http.Request) {
	var companies []models.TMCCompany
	database.DB.Find(&companies)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(companies)
}

func getStatuses(w http.ResponseWriter, r *http.Request) {
	var statuses []models.TMCStatus
	database.DB.Find(&statuses)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
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

	// ✅ Cek Data Duplikat
	var existing models.TMCAttendance
	result := database.DB.Where("name = ? AND company = ? AND status = ?", form.Name, form.Company, form.Status).First(&existing)

	if result.RowsAffected > 0 {
		http.Error(w, "Already exists.", http.StatusBadRequest)
		return
	}

	// ✅ Generate Unique Numeric ID (SEBELUM SAVE)
	uniqueID, err := generateUniqueNumericID()
	if err != nil {
		http.Error(w, "Failed to generate unique ID", http.StatusInternalServerError)
		return
	}
	form.UniqueID = uniqueID

	// ✅ Simpan Data Jika Belum Ada
	createResult := database.DB.Create(&form)
	if createResult.Error != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Data Saved: %+v\n", form)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully submitted!",
		"id":      encryptID(form.ID),
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
		http.Error(w, "Data not found", http.StatusNotFound)
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
    // Ambil unique_id dari parameter URL
    params := mux.Vars(r)
    uniqueID := params["id"]

    var attendance models.TMCAttendance

    // Cari data peserta berdasarkan unique_id
    result := database.DB.Where("unique_id = ?", uniqueID).First(&attendance)
    if result.Error != nil {
        http.Error(w, "Data not found", http.StatusNotFound)
        return
    }

    // Validasi apakah peserta sudah hadir
    if attendance.AttendedAt.Valid {
        http.Error(w, "Already attended", http.StatusBadRequest)
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
        http.Error(w, "Failed to save attendance", http.StatusInternalServerError)
        return
    }

    // Berikan respon sukses
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Attendance saved successfully",
        "data":    attendance,
    })
}

func allAttendeesHandler(w http.ResponseWriter, r *http.Request) {
    var attendees []models.TMCAttendance

    // Ambil semua data peserta
    result := database.DB.Order("created_at desc").Find(&attendees)
    if result.Error != nil {
        http.Error(w, "Failed to fetch attendees", http.StatusInternalServerError)
        return
    }

    // Berikan respon sukses
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message":   "Attendees fetched successfully",
        "attendees": attendees,
    })
}





