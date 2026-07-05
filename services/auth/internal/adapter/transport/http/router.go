package http

import (
	"auth/internal/adapter/middleware"
	"auth/internal/adapter/transport/http/handler"
	"auth/internal/ports"
	"net/http"
)

func SetupRouter(authHandler *handler.AuthHandler, teamHandler *handler.TeamHandler, tokenMgr ports.TokenManager) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)

	mux.Handle("POST /api/v1/teams", middleware.AuthMiddleware(tokenMgr)(http.HandlerFunc(teamHandler.CreateTeam)))

	return mux
}
