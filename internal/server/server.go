package server

import (
	"catbox-scanner-master/internal/config"
	"catbox-scanner-master/internal/database"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	db         *database.Database
	httpServer *http.Server
}

func NewServer(db *database.Database) *Server {
	return &Server{
		db: db,
	}
}

func (s *Server) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/add", s.handleAddEntry).Methods("POST")
	r.HandleFunc("/count", s.handleGetCount).Methods("GET")

	addr := fmt.Sprintf(":%s", config.AppConfig.Port)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Printf("Server started on port %s. Waiting for requests...", config.AppConfig.Port)
	if err := s.httpServer.ListenAndServe(); err != nil {
		log.Printf("Server stopped: %v", err)
	}
}

func (s *Server) Stop() {
	log.Println("Graceful shutdown requested.")
	s.httpServer.Shutdown(context.Background())
}

func (s *Server) handleAddEntry(w http.ResponseWriter, r *http.Request) {
	authKey := r.URL.Query().Get("auth")
	if authKey != config.AppConfig.AuthKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		ID  string `json:"id"`
		Ext string `json:"ext"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	if payload.ID == "" || payload.Ext == "" {
		http.Error(w, "Missing id or ext", http.StatusBadRequest)
		return
	}

	s.db.InsertEntry(payload.ID, payload.Ext)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Entry added: ID=%s, EXT=%s", payload.ID, payload.Ext)
}

func (s *Server) handleGetCount(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.GetTotalRows()
	if err != nil {
		http.Error(w, "Failed to retrieve total rows", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%d", rows)
}
