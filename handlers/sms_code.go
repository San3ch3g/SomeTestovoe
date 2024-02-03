// sms_code.go
package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SMSCodeHandler struct {
	DB *sql.DB
}

func NewSMSCodeHandler(db *sql.DB) *SMSCodeHandler {
	return &SMSCodeHandler{
		DB: db,
	}
}

func (h *SMSCodeHandler) Generate(c *gin.Context) {
	phoneNumber := c.PostForm("phoneNumber")
	idempotencyKey := c.PostForm("idempotencyKey")
	createdAt := time.Now()

	if phoneNumber == "" {
		log.Println("Phone number is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is required"})
		return
	}

	code := generateRandomCode()

	_, err := h.DB.Exec(`
		INSERT INTO sms_codes (code, phone_number, idempotency_key, created_at)
		VALUES (?, ?, ?, ?);
	`, code, phoneNumber, idempotencyKey, createdAt)

	if err != nil {
		log.Printf("Error inserting SMS code into the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save SMS code"})
		return
	}

	log.Printf("SMS code saved successfully: %s", code)

	c.JSON(http.StatusOK, gin.H{"message": "SMS code generated successfully"})
}

func (h *SMSCodeHandler) Verify(c *gin.Context) {
	idempotencyKey := c.PostForm("idempotencyKey")
	userCode := c.PostForm("code")

	var savedCode SMSCode
	var createdAtStr string

	err := h.DB.QueryRow(`
		SELECT id, code, phone_number, idempotency_key, created_at
		FROM sms_codes
		WHERE idempotency_key = ?;
	`, idempotencyKey).Scan(&savedCode.ID, &savedCode.Code, &savedCode.PhoneNumber, &savedCode.IdempotencyKey, &createdAtStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error fetching SMS code: %v", err)})
		return
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error parsing created_at: %v", err)})
		return
	}

	if time.Since(createdAt) > 5*time.Minute {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Time is gone"})
		return
	}

	savedCode.CreatedAt = createdAt

	if userCode != savedCode.Code {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid SMS code. User provided: %s, Saved code: %s", userCode, savedCode.Code),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SMS code verified successfully"})
}

func generateRandomCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%04d", rand.Intn(10000))
}

// Структура для хранения информации о SMS-коде
type SMSCode struct {
	ID             int       `json:"id"`
	Code           string    `json:"code"`
	PhoneNumber    string    `json:"phoneNumber"`
	IdempotencyKey string    `json:"idempotencyKey"`
	CreatedAt      time.Time `json:"createdAt"`
}
