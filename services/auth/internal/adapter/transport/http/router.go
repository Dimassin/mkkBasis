package http

import (
	"net/http"
	"time"

	"auth/internal/adapter/middleware"
	"auth/internal/adapter/transport/http/handler"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(authHandler *handler.AuthHandler) http.Handler {
	mux := http.NewServeMux()
	limiter := middleware.NewRateLimiter(100, time.Minute)

	mux.Handle("GET /metrics", promhttp.Handler())
	mux.Handle("POST /api/v1/register", limiter.Middleware(http.HandlerFunc(authHandler.Register)))
	mux.Handle("POST /api/v1/login", limiter.Middleware(http.HandlerFunc(authHandler.Login)))

	return middleware.Metrics(mux)
}
