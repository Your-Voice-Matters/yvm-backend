package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lpernett/godotenv"
	"github.com/rs/cors"

	"yvm-backend/controllers/polls"
	"yvm-backend/controllers/users"
	"yvm-backend/controllers/votes"
	"yvm-backend/helper"
	"yvm-backend/middleware"
	"yvm-backend/services"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func verifyLoginToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value("options").(jwt.MapClaims)
	if !ok {
		http.Error(w, "No token claims found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(claims)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"message": "Logged out successfully"})
}

// handler for the '/' route - used for health checks
func dflt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "All Good!"})
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	services.InitSupabase()

	c := cors.New(cors.Options{
		// AllowedOrigins: []string{"http://localhost:5173"},
		AllowedOrigins: []string{"https://yourvoicematters.netlify.app"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	http.HandleFunc("/", dflt)

	http.HandleFunc("/login", users.Login)
	http.HandleFunc("/signup", users.Signup)

	http.HandleFunc("/get-token-details", helper.ChainMiddleware(verifyLoginToken, middleware.VerifyTokenMiddleware))
	http.HandleFunc("/my-polls", helper.ChainMiddleware(polls.GetMyPolls, middleware.VerifyTokenMiddleware))
	http.HandleFunc("/polls-i-participated-in", helper.ChainMiddleware(polls.GetPollsIParticipatedIn, middleware.VerifyTokenMiddleware))
	http.HandleFunc("/most-popular-polls", polls.GetMostPopularPolls)
	http.HandleFunc("/create-poll", helper.ChainMiddleware(polls.CreatePoll, middleware.VerifyTokenMiddleware))
	http.HandleFunc("/has-voted", helper.ChainMiddleware(votes.HasVoted, middleware.VerifyTokenMiddleware))
	http.HandleFunc("/cast-vote", helper.ChainMiddleware(votes.CastVote, middleware.VerifyTokenMiddleware))
	http.HandleFunc("/poll-details", polls.GetPollDetails)
	http.HandleFunc("/logout", helper.ChainMiddleware(logoutHandler, middleware.VerifyTokenMiddleware))

	fmt.Println("Server started at port:" + port)
	handlerCORS := c.Handler(http.DefaultServeMux)
	log.Fatal(http.ListenAndServe(":"+port, handlerCORS))
}
