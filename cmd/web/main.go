package main

import (
	"log"
	"net/http"
	"ws/internal/handlers"
)

func main() {

	mux := routes()

	log.Println("Starting channel listener 🚀😎🤘")
	go handlers.ListenToTheWsChannelActivity()
	
	log.Println("Starting web server on port 8080")

	_ = http.ListenAndServe(":8080", mux)
}