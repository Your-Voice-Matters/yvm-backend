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

	// Clear JWT cookie (HttpOnly)
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	})

	// Clear CSRF cookie (assumes it's accessible via JS, so no HttpOnly)
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	})

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
		AllowedOrigins:   []string{"https://yvm-frontend1.vercel.app"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-CSRF-Token"},
		AllowCredentials: true,
	})

	http.HandleFunc("/", dflt)

	http.HandleFunc("/login", users.Login)
	http.HandleFunc("/signup", users.Signup)

	http.HandleFunc("/get-token-details", helper.ChainMiddleware(verifyLoginToken, middleware.VerifyTokenMiddleware))
	http.HandleFunc("/my-polls", helper.ChainMiddleware(polls.GetMyPolls, middleware.VerifyTokenMiddleware, middleware.VerifyCSRFtokens))
	http.HandleFunc("/polls-i-participated-in", helper.ChainMiddleware(polls.GetPollsIParticipatedIn, middleware.VerifyTokenMiddleware, middleware.VerifyCSRFtokens))
	http.HandleFunc("/most-popular-polls", polls.GetMostPopularPolls)
	http.HandleFunc("/create-poll", helper.ChainMiddleware(polls.CreatePoll, middleware.VerifyTokenMiddleware, middleware.VerifyCSRFtokens))
	http.HandleFunc("/has-voted", helper.ChainMiddleware(votes.HasVoted, middleware.VerifyTokenMiddleware, middleware.VerifyCSRFtokens))
	http.HandleFunc("/cast-vote", helper.ChainMiddleware(votes.CastVote, middleware.VerifyTokenMiddleware, middleware.VerifyCSRFtokens))
	http.HandleFunc("/poll-details", polls.GetPollDetails)
	http.HandleFunc("/logout", helper.ChainMiddleware(logoutHandler, middleware.VerifyTokenMiddleware, middleware.VerifyCSRFtokens))

	fmt.Println("Server started at http://localhost:" + port)
	handlerCORS := c.Handler(http.DefaultServeMux)
	log.Fatal(http.ListenAndServe(":"+port, handlerCORS))
}
