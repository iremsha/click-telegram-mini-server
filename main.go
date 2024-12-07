package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}
	
	connectDatabase(cfg.MongoDSN, cfg.MongoDatabaseName)
	go startBot(cfg.TelegramToken)

	r := mux.NewRouter()
	apiRouter := r.PathPrefix("/").Subrouter()

	apiRouter.Use(enableCORS)

	apiRouter.HandleFunc("/user", getPlayerHandler).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/{telegramId}/energy", updateEnergyHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/users/{telegramId}/points", updatePointsHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/users/{telegramId}/skills", getSkillsHandler).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/{telegramId}/skills", updateSkillsHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/leaderboard", getLeaderboardHandler).Methods("GET", "OPTIONS")

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
