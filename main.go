package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

var templates *template.Template

func init() {
	InitializeGameData()
	InitDatabase()

	// Parse HTML templates
	templates = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	// Static file serving
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/assets/teams/", http.StripPrefix("/assets/teams/", http.FileServer(http.Dir("assets/teams"))))

	// Main routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/game", gameHandler)

	// API routes
	http.HandleFunc("/api/start-game", startGameHandler)
	http.HandleFunc("/api/guess", guessHandler)
	http.HandleFunc("/api/autocomplete", autocompleteHandler)
	http.HandleFunc("/api/submit-score", submitScoreHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	leaderboard, err := GetFormattedTop10()
	if err != nil {
		log.Printf("Error getting leaderboard: %v", err)
		leaderboard = []FormattedLeaderboardEntry{}
	}

	data := struct {
		Leaderboard []FormattedLeaderboardEntry
	}{
		Leaderboard: leaderboard,
	}

	err = templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "game.html", nil)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}

type StartGameResponse struct {
	SessionID string `json:"sessionId"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
}

type GuessRequest struct {
	SessionID  string `json:"sessionId"`
	PlayerName string `json:"playerName"`
}

type GuessResponse struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message,omitempty"`
	Correct    bool         `json:"correct"`
	Comparison *GuessResult `json:"comparison,omitempty"`
	Score      int          `json:"score"`
	TimeLeft   int          `json:"timeLeft"`
	GameOver   bool         `json:"gameOver"`
	NextPlayer bool         `json:"nextPlayer"`
}

type AutocompleteResponse struct {
	Players []string `json:"players"`
}

type SubmitScoreRequest struct {
	SessionID string `json:"sessionId"`
	Username  string `json:"username"`
}

type SubmitScoreResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func startGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	session, err := CreateNewSession()
	if err != nil {
		log.Printf("Error creating session: %v", err)
		response := StartGameResponse{
			Success: false,
			Message: "Failed to create game session",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := StartGameResponse{
		SessionID: session.SessionID,
		Success:   true,
	}

	json.NewEncoder(w).Encode(response)
}

func guessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req GuessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := GuessResponse{
			Success: false,
			Message: "Invalid request format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.SessionID == "" || req.PlayerName == "" {
		response := GuessResponse{
			Success: false,
			Message: "SessionID and PlayerName are required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	session, exists := GetSession(req.SessionID)
	if !exists {
		response := GuessResponse{
			Success: false,
			Message: "Session not found",
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	if IsGameOver(session) {
		response := GuessResponse{
			Success:  false,
			Message:  "Game session has ended",
			GameOver: true,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	result, err := ValidateGuess(session, req.PlayerName)
	if err != nil {
		log.Printf("Error validating guess: %v", err)
		response := GuessResponse{
			Success: false,
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	isCorrect := CheckCorrectGuess(session, req.PlayerName)
	timeLeft := GetTimeRemaining(session)

	response := GuessResponse{
		Success:    true,
		Correct:    isCorrect,
		Comparison: result,
		Score:      session.Score,
		TimeLeft:   timeLeft,
		GameOver:   IsGameOver(session),
		NextPlayer: isCorrect || GetCurrentPlayerGuesses(session) >= 6,
	}

	json.NewEncoder(w).Encode(response)
}

func autocompleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query().Get("query")
	if query == "" {
		// Return empty array if no query provided
		response := AutocompleteResponse{
			Players: []string{},
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get all player names if query is too short, otherwise filter
	var players []string
	if len(strings.TrimSpace(query)) < 2 {
		// For very short queries, return first 50 names to avoid overwhelming the UI
		allNames := GetAllPlayerNames()
		if len(allNames) > 50 {
			players = allNames[:50]
		} else {
			players = allNames
		}
	} else {
		// Filter by query with a reasonable limit
		players = FilterPlayersByName(query, 50)
	}

	if players == nil {
		players = []string{}
	}

	response := AutocompleteResponse{
		Players: players,
	}

	json.NewEncoder(w).Encode(response)
}

func submitScoreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req SubmitScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := SubmitScoreResponse{
			Success: false,
			Message: "Invalid request format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.SessionID == "" || req.Username == "" {
		response := SubmitScoreResponse{
			Success: false,
			Message: "SessionID and Username are required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate and sanitize username
	username := SanitizeInput(req.Username)
	if len(username) == 0 || len(username) > 50 {
		response := SubmitScoreResponse{
			Success: false,
			Message: "Username must be between 1 and 50 characters",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get session
	session, exists := GetSession(req.SessionID)
	if !exists {
		response := SubmitScoreResponse{
			Success: false,
			Message: "Session not found",
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate session is completed or game over
	if !session.IsCompleted && !IsGameOver(session) {
		response := SubmitScoreResponse{
			Success: false,
			Message: "Game session is still active",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Calculate final score if not already done
	finalScore := session.Score
	if !session.IsCompleted {
		finalScore = session.CalculateFinalScore()
	}

	// Submit score to leaderboard
	err := AddToLeaderboardFromSession(username, session)
	if err != nil {
		log.Printf("Error adding score to leaderboard: %v", err)
		response := SubmitScoreResponse{
			Success: false,
			Message: "Failed to save score to leaderboard",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Score submitted for user %s: %d points (session %s)", username, finalScore, req.SessionID)

	response := SubmitScoreResponse{
		Success: true,
		Message: "Score submitted successfully",
	}

	json.NewEncoder(w).Encode(response)
}
