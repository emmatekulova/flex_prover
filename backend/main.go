package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"flex-prover-backend/handler"
	"flex-prover-backend/service"
)

func main() {
	// Load .env.whale from project root, fall back to local .env
	for _, f := range []string{"../.env.whale", ".env.whale", ".env"} {
		if err := godotenv.Load(f); err == nil {
			log.Printf("loaded env from %s", f)
			break
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	hedera, err := service.NewHederaService()
	if err != nil {
		log.Fatalf("failed to init Hedera service: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /nonce",             handler.Nonce())
	mux.HandleFunc("POST /verify",           handler.Verify(hedera))
	mux.HandleFunc("POST /prepare-associate", handler.PrepareAssociate(hedera))
	mux.HandleFunc("POST /submit-associate", handler.SubmitAssociate(hedera))

	origin := os.Getenv("ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:3000"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      withCORS(origin, mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("whale backend listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func withCORS(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
