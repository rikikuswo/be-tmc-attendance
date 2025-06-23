package models

import (
	"database/sql"
	"time"
)

type TMCStatus struct {
	ID        	uint      `gorm:"primaryKey" json:"id"`
	StatusName string    `gorm:"size:100" json:"status_name"`
	Status    	string    `gorm:"size:50" json:"status"`
	CreatedAt 	time.Time `json:"created_at"`
	UpdatedAt 	time.Time `json:"updated_at"`
	DeletedAt 	sql.NullTime `json:"updated_at"`
}
