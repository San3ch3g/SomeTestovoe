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

func NewGoogleLoginHandler(oauthConfig *oauth2.Config, userTokenFunc func(email string, idempotencyKey string) string, db *sql.DB) *GoogleLoginHandler {
	return &GoogleLoginHandler{
		oauthConfig:   oauthConfig,
		userTokenFunc: userTokenFunc,
		db:            db,
	}
}

func (h *GoogleLoginHandler) Login(c *gin.Context) {
	idempotencyKey := c.PostForm("idempotencyKey")

	url := h.oauthConfig.AuthCodeURL(idempotencyKey, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Println(url)
	http.Redirect(c.Writer, c.Request, url, http.StatusTemporaryRedirect)
}

func (h *GoogleLoginHandler) Callback(c *gin.Context) {
	idempotencyKey := c.Query("state")
	oauthCode := c.Query("code")

	token, err := h.oauthConfig.Exchange(context.Background(), oauthCode)
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to insert user into the database: %s", err.Error())})
			return
		}

		log.Printf("User inserted into the database: %s", email)
	}

	newToken := h.userTokenFunc(email, idempotencyKey)
	log.Printf("Token generated successfully for user: %s", email)
	c.JSON(http.StatusOK, gin.H{"token": newToken})
}
