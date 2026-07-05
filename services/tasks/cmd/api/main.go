package main

import (
	"context"
	"log"
	"net/http"
	"os"

	rediscache "tasks/internal/adapter/cache/redis"
	"tasks/internal/adapter/httpserver"
	"tasks/internal/adapter/jwtadapter"
	mysqldb "tasks/internal/adapter/repository/mysql"
	httptransport "tasks/internal/adapter/transport/http"
	"tasks/internal/adapter/transport/http/handler"
	"tasks/internal/ports"
	"tasks/internal/usecase"

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

	var taskListCache ports.TaskListCache
	var redisCleanup func()
	redisClient, err := rediscache.NewClient(context.Background())
	if err != nil {
		log.Println("Redis unavailable, task list cache disabled:", err)
	} else {
		taskListCache = rediscache.NewTaskListCache(redisClient)
		redisCleanup = func() { redisClient.Close() }
	}

	taskRepo := mysqldb.NewTaskRepository(db)
	teamRepo := mysqldb.NewTeamRepository(db)
	reportRepo := mysqldb.NewReportRepository(db)
	jwtValidator := jwtadapter.NewJWTValidator(jwtSecret)

	taskUsecase := usecase.NewTaskUsecase(taskRepo, teamRepo, taskListCache)
	reportUsecase := usecase.NewReportUsecase(reportRepo)
	taskHandler := handler.NewTaskHandler(taskUsecase)
	reportHandler := handler.NewReportHandler(reportUsecase)

	router := httptransport.SetupRouter(taskHandler, reportHandler, jwtValidator)

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8082"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	if err := httpserver.Run("Tasks server", server, func() {
		if redisCleanup != nil {
			redisCleanup()
		}
		db.Close()
	}); err != nil {
		log.Fatal(err)
	}
}
