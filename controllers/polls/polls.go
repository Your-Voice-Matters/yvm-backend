package polls

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

func GetMyPolls(w http.ResponseWriter, r *http.Request) {
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
	username := claims["username"].(string)
	resp := services.Client.Rpc("pollsICreated", "exact", map[string]string{"uname": username})
	var polls []map[string]any
	err := json.Unmarshal([]byte(resp), &polls)
	if err != nil {
		logger.Error("Error unmarshalling the response", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}

	json.NewEncoder(w).Encode(polls)
}

func CreatePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
	username := claims["username"].(string)

	var poll structs.PollObj
	err := json.NewDecoder(r.Body).Decode(&poll)
	if err != nil {
		http.Error(w, "An error occured while creating the poll. Please check again.", http.StatusBadRequest)
		return
	}
	poll.Creator = username
	_, _, err = services.Client.From("polls").Insert(poll, false, "", "", "exact").Execute()
	if err != nil {
		logger.Error("Error running the query for creating a poll", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Poll created successfully"})
}

func GetPollDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	pollID := r.URL.Query().Get("pollid")
	if pollID == "" {
		http.Error(w, "pollid is required", http.StatusBadRequest)
		return
	}
	resp, num, err := services.Client.From("polls").Select("*", "exact", false).Eq("id", pollID).Execute()
	if err != nil {
		logger.Error("Error running the query for getting the poll details", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	if num == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Poll not found"})
		return
	}
	var polls []structs.PollObj
	err = json.Unmarshal(resp, &polls)
	if err != nil {
		logger.Error("Error unmarshalling the response", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	resp = []byte(services.Client.Rpc("polloptioncounts", "exact", map[string]string{"pid": pollID}))
	var optionCounts []map[string]any
	err = json.Unmarshal(resp, &optionCounts)
	if err != nil {
		logger.Error("Error unmarshalling the response for option counts", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"poll": polls[0], "option_votes": optionCounts})
}

func GetPollsIParticipatedIn(w http.ResponseWriter, r *http.Request) {
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
	username := claims["username"].(string)
	resp := services.Client.Rpc("getPollsIParticipatedIn", "exact", map[string]string{"uname": username})
	var polls []map[string]any
	err := json.Unmarshal([]byte(resp), &polls)
	if err != nil {
		logger.Error("Error unmarshalling the response", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	json.NewEncoder(w).Encode(polls)
}

func GetMostPopularPolls(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	resp := services.Client.Rpc("mostPopularPolls", "exact", map[string]any{})
	var polls []map[string]any
	err := json.Unmarshal([]byte(resp), &polls)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error("Error unmarshalling the response", slog.String("error", err.Error()))
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	json.NewEncoder(w).Encode(polls)
}
