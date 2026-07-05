package http

import (
	"net/http"
	"time"

	"teams/internal/adapter/middleware"
	"teams/internal/adapter/transport/http/handler"
	"teams/internal/ports"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(teamHandler *handler.TeamHandler, tokenValidator ports.TokenValidator) http.Handler {
	mux := http.NewServeMux()
	limiter := middleware.NewRateLimiter(100, time.Minute)

	mux.Handle("GET /metrics", promhttp.Handler())

	mux.Handle("POST /api/v1/teams", chain(limiter, tokenValidator, teamHandler.CreateTeam))
	mux.Handle("GET /api/v1/teams", chain(limiter, tokenValidator, teamHandler.GetUserTeams))
	mux.Handle("POST /api/v1/teams/{id}/invite", chain(limiter, tokenValidator, teamHandler.InviteMember))

	return middleware.Metrics(mux)
}

func chain(limiter *middleware.RateLimiter, tokenValidator ports.TokenValidator, handler func(http.ResponseWriter, *http.Request)) http.Handler {
	return middleware.AuthMiddleware(tokenValidator)(limiter.Middleware(http.HandlerFunc(handler)))
}
