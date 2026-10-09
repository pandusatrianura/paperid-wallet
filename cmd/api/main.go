package main

import (
	"log"
	"os"

	"paperid-wallet/internal/app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	application := app.BuildApp(app.Config{Port: port})
	if err := application.Run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
