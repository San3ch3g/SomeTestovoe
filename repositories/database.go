package repositories

import (
	"database/sql"
	"log"
)

type DatabaseRepository struct {
	db *sql.DB
}

func NewDatabaseRepository(db *sql.DB) *DatabaseRepository {
	return &DatabaseRepository{db: db}
}

func (repo *DatabaseRepository) InitTables() {
	repo.initUsersTable()
	repo.initSmsCodesTable()
}

func (repo *DatabaseRepository) initUsersTable() {
	createTableQuery := `
        CREATE TABLE IF NOT EXISTS users (
            id INT AUTO_INCREMENT PRIMARY KEY,
            google_id VARCHAR(255) NOT NULL,
            email VARCHAR(255) NOT NULL,
            name VARCHAR(255),
            unique_value VARCHAR(255) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        ) DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
    `
	_, err := repo.db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func (repo *DatabaseRepository) initSmsCodesTable() {
	createTableQuery := `
        CREATE TABLE IF NOT EXISTS sms_codes (
            id INT AUTO_INCREMENT PRIMARY KEY,
            code VARCHAR(6) NOT NULL,
            phone_number VARCHAR(15) NOT NULL,
            idempotency_key VARCHAR(255) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `
	_, err := repo.db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}
