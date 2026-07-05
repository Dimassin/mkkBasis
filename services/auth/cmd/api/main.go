package main

import (
	"auth/internal/adapter/repository/mysql"
	"auth/internal/adapter/transport/http/handler"
	"auth/internal/usecase"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"auth/internal/adapter/hasher"
	"auth/internal/adapter/jwtadapter"
	httptransport "auth/internal/adapter/transport/http"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	jwtSecret := os.Getenv("JWT_SECRET")

	// для локальной среды
	if dbHost == "" {
		dbHost = "mysql"
	}
	if dbPort == "" {
		dbPort = "3306"
	}
	if dbUser == "" {
		dbUser = "myuser"
	}
	if dbPassword == "" {
		dbPassword = "mypassword"
	}
	if dbName == "" {
		dbName = "auth"
	}
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}

	dsn := dbUser + ":" + dbPassword + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	userRepo := mysql.NewUserRepository(db)
	passwordHasher := hasher.NewBcryptHasher()
	accessTTL := 15 * time.Minute
	jwtManager := jwtadapter.NewJWTManager(jwtSecret, accessTTL)
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtManager, passwordHasher)

	authHandler := handler.NewAuthHandler(authUsecase)
	router := httptransport.SetupRouter(authHandler)
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Println("Server starting on port", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed:", err)
	}
}
