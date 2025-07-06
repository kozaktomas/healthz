package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pires/go-proxyproto"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if os.Getenv("DEBUG") == "true" {
			log.Printf("Health check endpoint hit")
		}
		fmt.Fprintln(w, "Ok")
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Ok")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Create a listener
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatalf("Error creating listener: %v", err)
	}

	// Wrap the listener with Proxy Protocol support
	proxyListener := &proxyproto.Listener{Listener: listener}
	defer proxyListener.Close()

	// Channel to listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("Server is running on port 8080...")
		if err := server.Serve(proxyListener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for termination signal
	<-stop
	fmt.Println("\nShutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server stopped gracefully")
}
