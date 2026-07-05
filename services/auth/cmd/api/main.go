package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"auth/internal/adapter/hasher"
	"auth/internal/adapter/httpserver"
	"auth/internal/adapter/jwtadapter"
	mysqldb "auth/internal/adapter/repository/mysql"
	httptransport "auth/internal/adapter/transport/http"
	"auth/internal/adapter/transport/http/handler"
	"auth/internal/usecase"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	jwtSecret := os.Getenv("JWT_SECRET")

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
	db, err := mysqldb.OpenDB(dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	userRepo := mysqldb.NewUserRepository(db)
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

	if err := httpserver.Run("Auth server", server, func() { db.Close() }); err != nil {
		log.Fatal(err)
	}
}
