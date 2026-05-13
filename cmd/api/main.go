package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/shubh-samay/api/internal/config"
	"github.com/shubh-samay/api/internal/db"
	"github.com/shubh-samay/api/internal/handlers"
	"github.com/shubh-samay/api/internal/panchang"
)

func main() {
	loadDotenv(".env")

	cfg := config.Load()
	panchang.Init(cfg.EphePath)

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSON(w, http.StatusOK, handlers.JSONMap{"status": "ok"})
	})
	mux.HandleFunc("GET /v1/panchang", handlers.GetPanchang)
	mux.HandleFunc("GET /v1/festivals", handlers.GetFestivals)
	mux.HandleFunc("GET /v1/lunar-days", handlers.GetLunarDays)
	mux.HandleFunc("GET /v1/muhurat", handlers.FindMuhurat)
	mux.HandleFunc("GET /v1/config", handlers.GetConfig(database))
	mux.HandleFunc("POST /v1/devices", handlers.RegisterDevice(database))
	mux.HandleFunc("PATCH /v1/admin/flags/{key}", adminAuth(cfg.AdminToken, handlers.UpdateFlag(database)))
	mux.HandleFunc("GET /v1/admin/flags", adminAuth(cfg.AdminToken, handlers.ListFlags(database)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Shubh Samay API listening on :%s", port)
	if err := http.ListenAndServe(":"+port, corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Admin-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func adminAuth(token string, next handlers.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Admin-Token") != token {
			handlers.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

func loadDotenv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}
