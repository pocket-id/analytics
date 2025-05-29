package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Set up HTTP handlers
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("POST /heartbeat", func(w http.ResponseWriter, r *http.Request) {
		HeartbeatHandler(db)(w, r)
	})

	http.HandleFunc("GET /stats", func(w http.ResponseWriter, r *http.Request) {
		StatsHandler(db)(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server starting on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
