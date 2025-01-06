package main

import (
	"catbox-scanner-master/internal/config"
	"catbox-scanner-master/internal/database"
	"catbox-scanner-master/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	db, err := database.InitDatabase()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	srv := server.NewServer(db)
	go srv.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Received shutdown signal, stopping server...")

	srv.Stop()
	log.Println("Web Server Stopped")
	db.Stop()
	log.Println("Database Worker Stopped")

	log.Println("Application stopped gracefully.")
}
