package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	deliveryHttp "paperid-wallet/internal/delivery/http"
	"paperid-wallet/internal/domain"
	"paperid-wallet/internal/repository/inmemory"
	"paperid-wallet/internal/usecase"
)

// Config holds runtime configuration for the application.
type Config struct {
	Port string
}

// Application encapsulates the wired components and the HTTP server.
type Application struct {
	server *http.Server
}

// BuildApp wires together dependencies following Clean Architecture and SOLID principles.
func BuildApp(cfg Config) *Application {
	// 1. Data Store / Repository Layer (with seed data)
	seedUsers := map[string]*domain.User{
		"U001": {ID: "U001", Name: "Alice", Balance: 5000.00},
		"U002": {ID: "U002", Name: "Bob", Balance: 150.00},
	}
	userRepo := inmemory.NewUserRepository(seedUsers)

	// 2. Use Case Layer
	disburseUseCase := usecase.NewDisbursementUseCase(userRepo)

	// 3. Delivery Layer (HTTP Handlers & Router)
	disburseHandler := deliveryHttp.NewDisburseHandler(disburseUseCase)
	router := deliveryHttp.NewRouter(disburseHandler)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &Application{server: server}
}

// Run starts the HTTP server with graceful shutdown support.
func (a *Application) Run() error {
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Server starting on port %s...", a.server.Addr)
		serverErrors <- a.server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil

	case sig := <-shutdown:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.server.Shutdown(ctx); err != nil {
			_ = a.server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
		log.Println("Server stopped gracefully")
		return nil
	}
}
