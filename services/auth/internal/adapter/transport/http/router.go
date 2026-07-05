package http

import (
	"net/http"

	"auth/internal/adapter/transport/http/handler"
)

func SetupRouter(authHandler *handler.AuthHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/register", authHandler.Register)

	return mux
}
