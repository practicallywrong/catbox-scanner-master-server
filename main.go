package main

import (
	"catbox-scanner-master/internal/config"
	"catbox-scanner-master/internal/database"
	"catbox-scanner-master/internal/server"
	"catbox-scanner-master/internal/service"
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

	linkChecker := service.NewLinkChecker(db)
	go linkChecker.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Received shutdown signal, stopping services...")

	linkChecker.Stop()
	srv.Stop()
	log.Println("Web Server Stopped")
	db.Stop()
	log.Println("Database Worker Stopped")
	linkChecker.Stop()
	log.Println("Checker Service Stopped")

	log.Println("Master Server gracefully")
}
