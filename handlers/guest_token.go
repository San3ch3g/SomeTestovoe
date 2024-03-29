package handlers

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const secretKey = "IWannaGoToXenous"

type GuestTokenHandler struct {
}

func NewGuestTokenHandler() *GuestTokenHandler {
	return &GuestTokenHandler{}
}

type ResponseForTokens struct {
	Error string `json:"error"`
	Token string `json:"token"`
}

type Header struct {
	Authorization string `json:"Authorization"`
}

// GenerateGuestToken godoc
//
//	@Summary		Generate guest token
//	@Description	Generating jwt token
//	@Tags			GuestToken
//	@Produce		json
//	@Success		201	{object}	ResponseForTokens	"Some Response"
//
//	@Failure		500	{object}	Response			"Error response"
//
//	@Router			/generate-guest-token [post]
func (h *GuestTokenHandler) Generate(c *gin.Context) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = uuid.New().String()
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseForTokens{Error: "error with generating jwt token"})
		return
	}

	c.JSON(http.StatusCreated, ResponseForTokens{Token: tokenString})
}
