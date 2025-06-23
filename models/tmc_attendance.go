package models

import (
	"database/sql"
	"time"
)

type TMCAttendance struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UniqueID  string    `gorm:"uniqueIndex"`
	Name      string    `gorm:"size:100" json:"name"`
	Company   string    `gorm:"size:100" json:"company"`
	Status    string    `gorm:"size:50" json:"status"`
	Period	  string	`gorm:"size:20" json:"period"`
	AttendedAt sql.NullTime `json:"attended_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt sql.NullTime `json:"updated_at"`
}
