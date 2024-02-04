package main

import (
	"database/sql"
	"fmt"
	"log"

	"ModuleForTestTask/handlers"
	"ModuleForTestTask/repositories"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	dbUsername         = "sql11681286"
	dbPassword         = "wD4a4d2TfY"
	dbHost             = "sql11.freemysqlhosting.net"
	dbName             = "sql11681286"
	googleClientID     = "462095042255-6ihrp0qc3n7hdarfjr6sul1ikeg2pnrn.apps.googleusercontent.com"
	googleClientSecret = "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"
	serverPort         = ":8080"
	secretKey          = "IWannaGoToXenous"
)

var db *sql.DB

func main() {
	initDataBase()
	defer db.Close()

	dbRepository := repositories.NewDatabaseRepository(db)
	dbRepository.InitTables()

	r := gin.Default()

	// Создаем хендлеры
	guestTokenHandler := handlers.NewGuestTokenHandler()
	authHandler := handlers.NewAuthHandler()
	smsCodeHandler := handlers.NewSMSCodeHandler(db)
	userTokenHandler := handlers.NewUserTokenHandler(secretKey)
	googleOauthConfig := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  "http://localhost:8080/google-callback",
		Scopes:       []string{"profile", "email"},
		Endpoint:     google.Endpoint,
	}

	googleLoginHandler := handlers.NewGoogleLoginHandler(googleOauthConfig, userTokenHandler.GenerateToken, db)

	r.POST("/generate-guest-token", guestTokenHandler.Generate)

	authGroup := r.Group("/")
	authGroup.Use(authHandler.Middleware())
	{
		authGroup.POST("/generate-sms-code", smsCodeHandler.Generate)
		authGroup.POST("/verify-sms-code", smsCodeHandler.Verify)
		authGroup.POST("/login-google", googleLoginHandler.Login)
		r.GET("/google-callback", googleLoginHandler.Callback)
	}

	if err := r.Run(serverPort); err != nil {
		log.Fatal(err)
	}
}

func initDataBase() {
	dbURI := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUsername, dbPassword, dbHost, dbName)
	var err error
	db, err = sql.Open("mysql", dbURI)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}
