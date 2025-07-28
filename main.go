package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var templates *template.Template

func init() {
	if err := InitializeGameData(); err != nil {
		log.Fatalf("Failed to initialize game data: %v", err)
	}

	if err := InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Printf("Warning: Could not parse templates: %v", err)
		templates = template.New("empty")
	}
}

func main() {

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.HandleFunc("/riot.txt", func(w http.ResponseWriter, r *http.Request) {
		riotCode := os.Getenv("RIOT_VERIFICATION_CODE")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(riotCode))
	})

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/game", gameHandler)

	http.HandleFunc("/api/start-game", startGameHandler)
	http.HandleFunc("/api/guess", guessHandler)
	http.HandleFunc("/api/autocomplete", autocompleteHandler)
	http.HandleFunc("/api/submit-score", submitScoreHandler)
	http.HandleFunc("/api/end-game", endGameHandler)
	http.HandleFunc("/api/config", configHandler)

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

	facileLeaderboard, err := GetFormattedLeaderboardByDifficulty(10, "facile")
	if err != nil {
		log.Printf("Error getting facile leaderboard: %v", err)
		facileLeaderboard = []FormattedLeaderboardEntry{}
	}

	moyenLeaderboard, err := GetFormattedLeaderboardByDifficulty(10, "moyen")
	if err != nil {
		log.Printf("Error getting moyen leaderboard: %v", err)
		moyenLeaderboard = []FormattedLeaderboardEntry{}
	}

	difficileLeaderboard, err := GetFormattedLeaderboardByDifficulty(10, "difficile")
	if err != nil {
		log.Printf("Error getting difficile leaderboard: %v", err)
		difficileLeaderboard = []FormattedLeaderboardEntry{}
	}

	difficultyInfo := GetDifficultyInfo()

	data := struct {
		FacileLeaderboard    []FormattedLeaderboardEntry
		MoyenLeaderboard     []FormattedLeaderboardEntry
		DifficileLeaderboard []FormattedLeaderboardEntry
		DifficultyInfo       map[string]map[string]interface{}
	}{
		FacileLeaderboard:    facileLeaderboard,
		MoyenLeaderboard:     moyenLeaderboard,
		DifficileLeaderboard: difficileLeaderboard,
		DifficultyInfo:       difficultyInfo,
	}

	err = templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}

func gameHandler(w http.ResponseWriter, r *http.Request) {

	difficulty := r.URL.Query().Get("difficulty")
	if difficulty == "" {
		difficulty = "difficile"
	}

	difficultyInfo := GetDifficultyInfo()

	data := struct {
		Difficulty     string
		DifficultyInfo map[string]map[string]interface{}
	}{
		Difficulty:     difficulty,
		DifficultyInfo: difficultyInfo,
	}

	err := templates.ExecuteTemplate(w, "game.html", data)
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
	Rank    int    `json:"rank,omitempty"`
}

type StartGameRequest struct {
	Difficulty string `json:"difficulty"`
}

func startGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req StartGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		response := StartGameResponse{
			Success: false,
			Message: "Invalid request format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	difficulty := req.Difficulty
	if difficulty == "" {
		difficulty = "difficile"
	}

	session, err := CreateNewSessionWithDifficulty(difficulty)
	if err != nil {
		log.Printf("Error creating session with difficulty %s: %v", difficulty, err)
		response := StartGameResponse{
			Success: false,
			Message: "Failed to create game session",
		}
		w.WriteHeader(http.StatusInternalServerError)
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

	if session.IsGameOver() {
		response := GuessResponse{
			Success:  false,
			Message:  "Game session has ended",
			GameOver: true,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	isCorrect := session.CheckCorrectGuess(req.PlayerName)

	result, err := ValidateGuess(session, req.PlayerName)
	if err != nil {
		log.Printf("Error validating guess: %v", err)

		errorMsg := err.Error()
		if strings.Contains(errorMsg, "player not found:") {
			errorMsg = "Ce joueur n'est pas reconnu"
		} else if strings.Contains(errorMsg, "player not in difficulty:") {
			errorMsg = "Ce joueur n'est pas dans cette difficulté"
		}

		response := GuessResponse{
			Success: false,
			Message: errorMsg,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	timeLeft := GetTimeRemaining(session)

	response := GuessResponse{
		Success:    true,
		Correct:    isCorrect,
		Comparison: result,
		Score:      session.Score,
		TimeLeft:   timeLeft,
		GameOver:   session.IsGameOver(),
		NextPlayer: isCorrect,
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
	sessionID := r.URL.Query().Get("sessionId")

	if query == "" {
		response := AutocompleteResponse{
			Players: []string{},
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	difficulty := "difficile"
	if sessionID != "" {
		if session, exists := GetSession(sessionID); exists {
			difficulty = session.Difficulty
		} else {
			log.Printf("Autocomplete: session %s not found, using default difficulty", sessionID)
		}
	} else {
		log.Printf("Autocomplete: no session ID provided, using default difficulty")
	}

	var players []string
	if len(strings.TrimSpace(query)) < 2 {
		allNames := FilterPlayersByNameAndDifficulty("", difficulty, 50)
		if len(allNames) > 50 {
			players = allNames[:50]
		} else {
			players = allNames
		}
	} else {
		players = FilterPlayersByNameAndDifficulty(query, difficulty, 50)
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

	if !session.IsCompleted && !session.IsGameOver() {
		response := SubmitScoreResponse{
			Success: false,
			Message: "Game session is still active",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	finalScore := session.Score
	if !session.IsCompleted {
		finalScore = session.CalculateFinalScore()
	}

	err := SubmitScoreByDifficulty(username, session, session.Difficulty)
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

	var totalDuration int
	if session.CompletionTime != nil {
		totalDuration = int(session.CompletionTime.Sub(session.StartTime).Seconds())
	} else {
		totalDuration = int(time.Since(session.StartTime).Seconds())
	}

	rank, err := GetPlayerRankByDifficulty(finalScore, totalDuration, session.Difficulty)
	if err != nil {
		log.Printf("Error calculating rank: %v", err)

		rank = 0
	}

	log.Printf("Score submitted for user %s: %d points (rank #%d) (session %s)", username, finalScore, rank, req.SessionID)

	response := SubmitScoreResponse{
		Success: true,
		Message: "Score submitted successfully",
		Rank:    rank,
	}

	json.NewEncoder(w).Encode(response)
}

type EndGameRequest struct {
	SessionID string `json:"sessionId"`
}

type EndGameResponse struct {
	Success      bool    `json:"success"`
	Message      string  `json:"message,omitempty"`
	MissedPlayer *Player `json:"missed_player,omitempty"`
}

func endGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req EndGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := EndGameResponse{
			Success: false,
			Message: "Invalid request format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.SessionID == "" {
		response := EndGameResponse{
			Success: false,
			Message: "SessionID is required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	session, exists := GetSession(req.SessionID)
	if !exists {
		response := EndGameResponse{
			Success: false,
			Message: "Session not found",
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	if !session.IsCompleted {
		session.CompleteSession()
		UpdateSession(session)
	}

	// Get the current target player that was missed (if any)
	var missedPlayer *Player
	if session.CurrentPlayerIndex < len(session.SelectedPlayers) {
		currentPlayer := session.SelectedPlayers[session.CurrentPlayerIndex]
		missedPlayer = &currentPlayer
	}

	response := EndGameResponse{
		Success:      true,
		Message:      "Game session ended successfully",
		MissedPlayer: missedPlayer,
	}

	json.NewEncoder(w).Encode(response)
}

type GameConfig struct {
	TotalGameTimeSeconds int `json:"totalGameTimeSeconds"`
	PlayersPerSession    int `json:"playersPerSession"`
}

type ConfigResponse struct {
	Success bool        `json:"success"`
	Config  *GameConfig `json:"config,omitempty"`
	Message string      `json:"message,omitempty"`
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		config := &GameConfig{
			TotalGameTimeSeconds: TotalGameTime,
			PlayersPerSession:    PlayersPerSession,
		}

		response := ConfigResponse{
			Success: true,
			Config:  config,
		}

		json.NewEncoder(w).Encode(response)
		return
	}

	response := ConfigResponse{
		Success: false,
		Message: "Only GET method is supported for configuration",
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(response)
}
