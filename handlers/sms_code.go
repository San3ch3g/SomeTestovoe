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

type GenerateRequest struct {
	PhoneNumber    string `json:"phoneNumber"`
	IdempotencyKey string `json:"idempotencyKey"`
}
type Response struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}
type VerifyRequest struct {
	IdempotencyKey string `json:"idempotencyKey"`
	Code           string `json:"code"`
}
type VerifyResponse struct{}

// Generate godoc
//
//	@Summary		Generating sms code
//	@Description	Generate sms code to verify phone number
//	@Tags			SmsCode
//	@Produce		json
//	@Param			input	body		GenerateRequest	true	"Some info for generating"
//	@Param			input	header		Header			true	"Jwt token"
//	@Success		201		{object}	Response		"Some Response"
//
//	@Failure		400		{object}	Response		"Error response"
//
//	@Router			/generate-sms-code [post]
func (h *SMSCodeHandler) Generate(c *gin.Context) {
	createdAt := time.Now()
	var generateRequest GenerateRequest
	if err := c.ShouldBindJSON(&generateRequest); err != nil {
		c.JSON(http.StatusBadRequest, Response{Error: "Error with reading requst"})
		return
	}
	if generateRequest.PhoneNumber == "" {
		log.Println("Phone number is required")
		c.JSON(http.StatusBadRequest, Response{Error: "Phone number is required"})
		return
	}

	code := generateRandomCode()

	_, err := h.DB.Exec(`
		INSERT INTO sms_codes (code, phone_number, idempotency_key, created_at)
		VALUES (?, ?, ?, ?);
	`, code, generateRequest.PhoneNumber, generateRequest.IdempotencyKey, createdAt)

	if err != nil {
		log.Printf("Error inserting SMS code into the database: %v", err)
		c.JSON(http.StatusInternalServerError, Response{Error: "Failed to save SMS code"})
		return
	}

	log.Printf("SMS code saved successfully: %s", code)

	c.JSON(http.StatusCreated, Response{Message: "SMS code generated successfully"})
}

// Verify godoc
//
//	@Summary		Verifying sms code
//	@Description	Verifying sms code to verify phone number
//	@Tags			SmsCode
//	@Produce		json
//	@Param			input	body		VerifyRequest	true	"Some info for verifying"
//	@Param			input	header		Header			true	"Jwt token"
//	@Success		200		{object}	Response		"Some Response"
//
//	@Failure		400		{object}	Response		"Error response"
//
//	@Router			/verify-sms-code [post]
func (h *SMSCodeHandler) Verify(c *gin.Context) {
	var verifyRequest VerifyRequest
	if err := c.ShouldBindJSON(&verifyRequest); err != nil {
		c.JSON(http.StatusBadRequest, Response{Error: "Error with reading requst"})
		return
	}

	var savedCode SMSCode
	var createdAtStr string

	err := h.DB.QueryRow(`
		SELECT id, code, phone_number, idempotency_key, created_at
		FROM sms_codes
		WHERE idempotency_key = ?;
	`, verifyRequest.IdempotencyKey).Scan(&savedCode.ID, &savedCode.Code, &savedCode.PhoneNumber, &savedCode.IdempotencyKey, &createdAtStr)

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

	if verifyRequest.Code != savedCode.Code {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid SMS code. User provided: %s, Saved code: %s", verifyRequest.Code, savedCode.Code),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SMS code verified successfully"})
}

func generateRandomCode() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%04d", rand.Intn(10000))
}

type SMSCode struct {
	ID             int       `json:"id"`
	Code           string    `json:"code"`
	PhoneNumber    string    `json:"phoneNumber"`
	IdempotencyKey string    `json:"idempotencyKey"`
	CreatedAt      time.Time `json:"createdAt"`
}
