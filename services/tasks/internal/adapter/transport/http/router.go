package http

import (
	"net/http"
	"time"

	"tasks/internal/adapter/middleware"
	"tasks/internal/adapter/transport/http/handler"
	"tasks/internal/ports"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(taskHandler *handler.TaskHandler, reportHandler *handler.ReportHandler, tokenValidator ports.TokenValidator) http.Handler {
	mux := http.NewServeMux()
	limiter := middleware.NewRateLimiter(100, time.Minute)

	mux.Handle("GET /metrics", promhttp.Handler())

	mux.Handle("POST /api/v1/tasks", chain(limiter, tokenValidator, taskHandler.CreateTask))
	mux.Handle("GET /api/v1/tasks", chain(limiter, tokenValidator, taskHandler.ListTasks))
	mux.Handle("PUT /api/v1/tasks/{id}", chain(limiter, tokenValidator, taskHandler.UpdateTask))
	mux.Handle("GET /api/v1/tasks/{id}/history", chain(limiter, tokenValidator, taskHandler.GetTaskHistory))
	mux.Handle("GET /api/v1/reports/team-stats", chain(limiter, tokenValidator, reportHandler.GetTeamStats))
	mux.Handle("GET /api/v1/reports/top-creators", chain(limiter, tokenValidator, reportHandler.GetTopCreatorsByTeam))
	mux.Handle("GET /api/v1/reports/invalid-assignees", chain(limiter, tokenValidator, reportHandler.GetInvalidAssigneeTasks))

	return middleware.Metrics(mux)
}

func chain(limiter *middleware.RateLimiter, tokenValidator ports.TokenValidator, handler func(http.ResponseWriter, *http.Request)) http.Handler {
	return middleware.AuthMiddleware(tokenValidator)(limiter.Middleware(http.HandlerFunc(handler)))
}
