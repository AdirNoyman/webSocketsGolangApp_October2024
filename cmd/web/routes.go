package main

import (
	"net/http"
	"ws/internal/handlers"

	"github.com/bmizerany/pat"
)

func routes() http.Handler {

	// Create a router
	mux := pat.New()

	mux.Get("/", http.HandlerFunc(handlers.Home))
	mux.Get("/ws", http.HandlerFunc(handlers.WebSocketEndPoint))

	return mux

}