package users

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"yvm-backend/helper"
	"yvm-backend/services"
	"yvm-backend/structs"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var creds structs.UserCreds
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, num, err := services.Client.From("usercreds").Select("username, password", "exact", false).Eq("username", creds.Username).Execute()
	if err != nil {
		logger.Error("Error running the query for getting the list of users", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	if num == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid credentials"})
		return
	}

	var users []structs.UserCreds
	err = json.Unmarshal(resp, &users)
	if err != nil {
		logger.Error("Error unmarshalling the response", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(users[0].Password), []byte(creds.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid credentials"})
		return
	}

	// Create JWT token
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"username": creds.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}).SignedString([]byte(os.Getenv("PASSPHRASE")))
	if err != nil {
		logger.Error("Error generating token", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}

	// Generate CSRF token
	csrfToken, err := helper.GenerateCSRFToken()
	if err != nil {
		http.Error(w, "Could not generate CSRF token", http.StatusInternalServerError)
		return
	}

	// Set JWT token cookie (HttpOnly, Secure should be true in production HTTPS)
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	// Set CSRF token cookie (accessible to JS)
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	// Send minimal JSON response (without token)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged in successfully", "username": creds.Username})
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var creds structs.UserCreds
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, num, err := services.Client.From("usercreds").Select("username", "exact", true).Eq("username", creds.Username).Execute()
	if err != nil {
		logger.Error("Error running the query for getting the list of users", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	if num > 0 {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"message": "Username already taken"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 14)
	if err != nil {
		logger.Error("Error hashing password", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "An unknown error occured"})
		return
	}
	_, _, err2 := services.Client.From("usercreds").Insert(map[string]any{
		"username": creds.Username,
		"password": string(hashedPassword),
	}, false, "", "", "exact").Execute()
	if err2 != nil {
		http.Error(w, "Error signing up user", http.StatusInternalServerError)
		logger.Error("Error signing up user", slog.String("error", err2.Error()))
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Signed up successfully"})
}
