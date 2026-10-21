package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed web/*
var webFiles embed.FS

func main() {
	if err := loadDotEnv(".env"); err != nil {
		log.Fatalf("could not load .env: %v", err)
	}
	client := newGeminiClient(os.Getenv("GEMINI_API_KEY"), env("GEMINI_MODEL", "gemini-2.5-flash-lite"))
	if client.key == "" {
		log.Println("WARNING: GEMINI_API_KEY is missing; the website will load, but AI chat will not work")
	}
	limiter := newLimiter()
	defer limiter.close()
	mux := http.NewServeMux()
	files, err := fs.Sub(webFiles, "web")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("GET /", http.FileServer(http.FS(files)))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("ok"))
	})
	mux.Handle("POST /api/chat", chatHandler(client, limiter))
	port := env("PORT", "8080")
	srv := &http.Server{Addr: ":" + port, Handler: logRequests(mux), ReadHeaderTimeout: 10 * time.Second}
	log.Printf("月下东坡 started at http://localhost:%s (model=%s)", port, client.model)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
