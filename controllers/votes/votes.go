package votes

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"

	"yvm-backend/services"
	"yvm-backend/structs"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func CastVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	claims, ok := r.Context().Value("options").(jwt.MapClaims)
	if !ok {
		logger.Error("No claims found in context")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	var vote structs.Vote
	vote.Votername = claims["username"].(string)
	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		logger.Error("Error decoding the request body", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	_, _, err = services.Client.From("votes").Insert(vote, false, "", "", "exact").Execute()
	if err != nil {
		logger.Error("Error running the query for casting the vote", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Vote cast successfully"})
}

func HasVoted(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	claims, ok := r.Context().Value("options").(jwt.MapClaims)
	if !ok {
		logger.Error("No claims found in context")
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	pollID := r.URL.Query().Get("pollid")
	if pollID == "" {
		http.Error(w, "pollid is required", http.StatusBadRequest)
		return
	}
	username := claims["username"].(string)
	resp, num, err := services.Client.From("votes").Select("*", "exact", false).Eq("pollid", pollID).Eq("votername", username).Execute()
	if err != nil {
		logger.Error("Error running the query for checking if user has voted", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	votes := []structs.Vote{}
	err = json.Unmarshal(resp, &votes)
	if err != nil {
		logger.Error("Error unmarshalling the response", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	if num > 0 {
		json.NewEncoder(w).Encode(map[string]any{"hasVoted": true, "chosenoption": votes[0].OptionID, "description": votes[0].Description})
	} else {
		json.NewEncoder(w).Encode(map[string]any{"hasVoted": false})
	}
}
