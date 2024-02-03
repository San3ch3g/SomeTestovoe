package repositories

import (
	"database/sql"
)

type SMSCodeRepository struct {
	DB *sql.DB
}

func NewSMSCodeRepository(db *sql.DB) *SMSCodeRepository {
	return &SMSCodeRepository{DB: db}
}
