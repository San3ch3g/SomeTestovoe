package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type GoogleLoginHandler struct {
	oauthConfig   *oauth2.Config
	userTokenFunc func(email string, idempotencyKey string) string
	db            *sql.DB
}

type LoginRequest struct {
	IdempotencyKey string `json:"idempotencyKey"`
}
type CallbackRequest struct {
	IdempotencyKey string `json:"idempotencyKey"`
	OauthCode      string `json:"code"`
}

func NewGoogleLoginHandler(oauthConfig *oauth2.Config, userTokenFunc func(email string, idempotencyKey string) string, db *sql.DB) *GoogleLoginHandler {
	return &GoogleLoginHandler{
		oauthConfig:   oauthConfig,
		userTokenFunc: userTokenFunc,
		db:            db,
	}
}

// Login godoc
//
//	@Summary		Loginig
//	@Description	Logining
//	@Tags			GoogleLogin
//	@Produce		json
//	@Param			input	body		LoginRequest	true	"Data for logining"
//	@Param			input	header		Header			true	"jwt token"
//	@Success		200		{object}	Response		"Some Response"
//
//	@Failure		400		{object}	Response		"Error response"
//
//	@Router			/login-google [post]
func (h *GoogleLoginHandler) Login(c *gin.Context) {
	var loginRequest LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, Response{Error: "Error with reading requst"})
		return
	}

	url := h.oauthConfig.AuthCodeURL(loginRequest.IdempotencyKey, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Println(url)
	http.Redirect(c.Writer, c.Request, url, http.StatusTemporaryRedirect)
}

// Callback godoc
//
//	@Summary		Callback
//	@Description	Callback after redirect
//	@Tags			GoogleLogin
//	@Produce		json
//	@Success		200	{object}	Response	"Some Response"
//
//	@Failure		400	{object}	Response	"Error response"
//	@Failure		500	{object}	Response	"Error response"
//
//	@Router			/google-callback [get]
func (h *GoogleLoginHandler) Callback(c *gin.Context) {
	var callbackRequest CallbackRequest
	if err := c.ShouldBindQuery(&callbackRequest); err != nil {
		c.JSON(http.StatusBadRequest, Response{Error: "Error with reading request"})
		return
	}

	token, err := h.oauthConfig.Exchange(context.Background(), callbackRequest.OauthCode)
	if err != nil {
		log.Printf("Failed to exchange OAuth code: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to exchange OAuth code: %s", err.Error())})
		return
	}

	client := h.oauthConfig.Client(context.Background(), token)
	userInfoResponse, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Failed to get user info from Google: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get user info from Google: %s", err.Error())})
		return
	}
	defer userInfoResponse.Body.Close()

	var userInfo map[string]interface{}
	err = json.NewDecoder(userInfoResponse.Body).Decode(&userInfo)
	if err != nil {
		log.Printf("Failed to decode user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to decode user info: %s", err.Error())})
		return
	}

	email, ok := userInfo["email"].(string)
	if !ok {
		log.Println("Failed to extract email from user info")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract email from user info"})
		return
	}

	var userID int
	err = h.db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)

	if err == sql.ErrNoRows {
		name, _ := userInfo["name"].(string)
		uniqueValue := uuid.New().String()

		_, err := h.db.Exec(`
			INSERT INTO users (google_id, email, name, unique_value)
			VALUES (?, ?, ?, ?);
		`, uniqueValue, email, name, uniqueValue)

		if err != nil {
			log.Printf("Failed to insert user into the database: %v", err)
			c.JSON(http.StatusInternalServerError, Response{Error: fmt.Sprintf("Failed to insert user into the database: %s", err.Error())})
			return
		}

		log.Printf("User inserted into the database: %s", email)
	}
	newToken := h.userTokenFunc(email, callbackRequest.IdempotencyKey)
	log.Printf("Token generated successfully for user: %s", email)
	c.JSON(http.StatusOK, ResponseForTokens{Token: newToken})
}
