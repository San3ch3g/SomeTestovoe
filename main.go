package main

import (
	"database/sql"
	"fmt"
	"log"

	docs "ModuleForTestTask/docs"
	"ModuleForTestTask/handlers"
	"ModuleForTestTask/repositories"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

//	@title			Testovoe v Xenous
//	@version		1.0
//	@description	API server for test task in Xenous
//	@host			localhost:8080
//	@BasePath		/

const (
	dbUsername         = "sql8685323"
	dbPassword         = "KBfhMcltzX"
	dbHost             = "sql8.freemysqlhosting.net"
	dbName             = "sql8685323"
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
	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

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
