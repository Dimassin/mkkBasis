package main

import (
	"log"
	"net/http"
	"os"

	"teams/internal/adapter/email"
	"teams/internal/adapter/httpserver"
	"teams/internal/adapter/jwtadapter"
	mysqldb "teams/internal/adapter/repository/mysql"
	httptransport "teams/internal/adapter/transport/http"
	"teams/internal/adapter/transport/http/handler"
	"teams/internal/usecase"

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

	teamRepo := mysqldb.NewTeamRepository(db)
	userRepo := mysqldb.NewUserRepository(db)
	jwtValidator := jwtadapter.NewJWTValidator(jwtSecret)
	emailService := email.NewCircuitBreakerService(email.NewMockService())

	teamUsecase := usecase.NewTeamUsecase(teamRepo, userRepo, emailService)
	teamHandler := handler.NewTeamHandler(teamUsecase)

	router := httptransport.SetupRouter(teamHandler, jwtValidator)

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8081"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	if err := httpserver.Run("Teams server", server, func() { db.Close() }); err != nil {
		log.Fatal(err)
	}
}
