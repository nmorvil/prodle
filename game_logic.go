package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	activeSessions = make(map[string]*GameSession)
	sessionMutex   sync.RWMutex
)

const (
	PlayersPerSession = 20
	TotalGameTime     = 120
)

func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func CreateNewSessionWithDifficulty(difficulty string) (*GameSession, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %v", err)
	}

	players, err := GetRandomPlayersByDifficulty(PlayersPerSession, difficulty)
	if err != nil {
		return nil, fmt.Errorf("failed to get random players for difficulty %s: %v", difficulty, err)
	}

	now := time.Now()
	session := &GameSession{
		SessionID:          sessionID,
		Difficulty:         difficulty,
		SelectedPlayers:    players,
		CurrentPlayerIndex: 0,
		Score:              0,
		StartTime:          now,
		Guesses:            make([]GuessResult, 0),
		IsCompleted:        false,
		CompletionTime:     nil,
	}

	sessionMutex.Lock()
	activeSessions[sessionID] = session
	sessionMutex.Unlock()

	log.Printf("Created new session %s with difficulty %s and %d players", sessionID, difficulty, len(players))

	return session, nil
}

func GetSession(sessionID string) (*GameSession, bool) {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()

	session, exists := activeSessions[sessionID]

	if !exists {
		log.Printf("Session %s not found. Active sessions: %d", sessionID, len(activeSessions))
	}

	return session, exists
}

func UpdateSession(session *GameSession) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	activeSessions[session.SessionID] = session
}

func (gs *GameSession) GetCurrentPlayer() *Player {
	if gs.CurrentPlayerIndex >= 0 && gs.CurrentPlayerIndex < len(gs.SelectedPlayers) {
		return &gs.SelectedPlayers[gs.CurrentPlayerIndex]
	}
	return nil
}

func (gs *GameSession) GetTotalElapsedTime() int {
	return int(time.Since(gs.StartTime).Seconds())
}

func (gs *GameSession) GetCurrentScore() int {
	elapsedSeconds := gs.GetTotalElapsedTime()
	totalWrongGuesses := 0

	for _, guess := range gs.Guesses {
		if !guess.IsCorrect {
			totalWrongGuesses++
		}
	}

	return CalculateGameScore(elapsedSeconds, totalWrongGuesses, gs.CurrentPlayerIndex)
}

func (gs *GameSession) CheckCorrectGuess(guessedPlayerName string) bool {
	targetPlayer := gs.GetCurrentPlayer()
	if targetPlayer == nil {
		return false
	}

	guessedPlayer, exists := GetPlayerByName(guessedPlayerName)
	if !exists {
		return false
	}

	return guessedPlayer.PlayerUsername == targetPlayer.PlayerUsername
}

func (gs *GameSession) MoveToNextPlayer() bool {
	gs.CurrentPlayerIndex++

	if gs.CurrentPlayerIndex >= len(gs.SelectedPlayers) {
		return false // Game completed
	}

	gs.Guesses = make([]GuessResult, 0)

	log.Printf("Session %s moved to player %d/%d",
		gs.SessionID, gs.CurrentPlayerIndex+1, len(gs.SelectedPlayers))

	return true
}

func (gs *GameSession) IsGameOver() bool {
	if gs.IsCompleted {
		return true
	}

	elapsedSeconds := gs.GetTotalElapsedTime()
	if elapsedSeconds >= TotalGameTime {
		return true
	}

	if gs.CurrentPlayerIndex >= len(gs.SelectedPlayers) {
		return true
	}

	return false
}

func (gs *GameSession) CalculateFinalScore() int {
	if !gs.IsCompleted {
		log.Printf("Game ending without completion for session %s at player %d/%d",
			gs.SessionID, gs.CurrentPlayerIndex+1, len(gs.SelectedPlayers))
	}

	completedPlayers := gs.CurrentPlayerIndex
	if gs.IsCompleted && completedPlayers > len(gs.SelectedPlayers) {
		completedPlayers = len(gs.SelectedPlayers)
	}

	if completedPlayers == len(gs.SelectedPlayers) {
		completionBonus := 10000 // Much bigger bonus for completing all 20 players
		gs.Score += completionBonus
		log.Printf("Session %s completed all players! Bonus: %d points", gs.SessionID, completionBonus)
	}

	return gs.Score
}

func (gs *GameSession) CompleteSession() {
	gs.IsCompleted = true
	now := time.Now()
	gs.CompletionTime = &now

	finalScore := gs.CalculateFinalScore()

	duration := int(time.Since(gs.StartTime).Seconds())
	completedPlayers := gs.CurrentPlayerIndex
	if completedPlayers > len(gs.SelectedPlayers) {
		completedPlayers = len(gs.SelectedPlayers)
	}

	log.Printf("Session %s completed. Players: %d/%d, Final Score: %d, Duration: %ds",
		gs.SessionID, completedPlayers, len(gs.SelectedPlayers), finalScore, duration)
}

func ValidateGuess(session *GameSession, guessedPlayerName string) (*GuessResult, error) {
	if session == nil {
		return nil, fmt.Errorf("session is nil")
	}

	if session.IsCompleted {
		return nil, fmt.Errorf("session already completed")
	}

	guessedPlayerName = SanitizeInput(guessedPlayerName)
	if valid, errMsg := ValidatePlayerGuess(guessedPlayerName); !valid {
		return nil, fmt.Errorf("invalid guess: %s", errMsg)
	}

	guessedPlayer, exists := GetPlayerByName(guessedPlayerName)
	if !exists {
		return nil, fmt.Errorf("player not found: %s", guessedPlayerName)
	}

	if !IsPlayerInDifficulty(guessedPlayer, session.Difficulty) {
		return nil, fmt.Errorf("player not in difficulty: %s", guessedPlayerName)
	}

	// Get current target player
	targetPlayer := session.GetCurrentPlayer()
	if targetPlayer == nil {
		return nil, fmt.Errorf("no current target player")
	}

	// Compare guess with target using detailed comparison
	comparisons := comparePlayersDetailed(*guessedPlayer, *targetPlayer)
	isCorrect := guessedPlayer.PlayerUsername == targetPlayer.PlayerUsername

	// Create guess result
	guessResult := GuessResult{
		GuessedPlayer: *guessedPlayer,
		TargetPlayer:  *targetPlayer,
		Timestamp:     time.Now(),
		Comparisons:   comparisons,
		IsCorrect:     isCorrect,
	}

	// Add guess to session
	session.Guesses = append(session.Guesses, guessResult)

	// Handle correct guess
	if isCorrect {
		session.handleCorrectGuess()
	} else {
		// Check if game should end after this guess (only for time limit)
		if session.IsGameOver() {
			// Time limit reached
			session.handleTimeLimit()
		}
	}

	// Update session in storage
	UpdateSession(session)

	return &guessResult, nil
}

// comparePlayersDetailed compares two players and returns detailed comparison results
func comparePlayersDetailed(guessed, target Player) map[string]ComparisonResult {
	comparisons := make(map[string]ComparisonResult)

	// Team comparison - exact match or league match (partial)
	if guessed.PlayerTeam == target.PlayerTeam {
		comparisons["team"] = ComparisonExact
	} else if guessed.PlayerLeague == target.PlayerLeague {
		comparisons["team"] = ComparisonPartial // Same league, different team
	} else {
		comparisons["team"] = ComparisonWrong
	}

	// League comparison - exact match only
	if guessed.PlayerLeague == target.PlayerLeague {
		comparisons["league"] = ComparisonExact
	} else {
		comparisons["league"] = ComparisonWrong
	}

	// Role comparison - exact match only
	if guessed.PlayerRole == target.PlayerRole {
		comparisons["role"] = ComparisonExact
	} else {
		comparisons["role"] = ComparisonWrong
	}

	// Country comparison - exact match or continent match (partial)
	if guessed.PlayerCountry == target.PlayerCountry {
		comparisons["country"] = ComparisonExact
	} else if guessed.PlayerCountryContinent == target.PlayerCountryContinent {
		comparisons["country"] = ComparisonPartial // Same continent, different country
	} else {
		comparisons["country"] = ComparisonWrong
	}

	// Age comparison - exact, higher, or lower
	if guessed.PlayerAge == target.PlayerAge {
		comparisons["age"] = ComparisonExact
	} else if guessed.PlayerAge > target.PlayerAge {
		comparisons["age"] = ComparisonHigher
	} else {
		comparisons["age"] = ComparisonLower
	}

	// Number of clubs comparison - exact, higher, or lower
	if guessed.NumberOfClubs == target.NumberOfClubs {
		comparisons["clubs"] = ComparisonExact
	} else if guessed.NumberOfClubs > target.NumberOfClubs {
		comparisons["clubs"] = ComparisonHigher
	} else {
		comparisons["clubs"] = ComparisonLower
	}

	// KDA ratio comparison - exact, higher, or lower (with tolerance for floating point)
	kdaTolerance := 0.01
	if abs(guessed.KDARatio-target.KDARatio) < kdaTolerance {
		comparisons["kda"] = ComparisonExact
	} else if guessed.KDARatio > target.KDARatio {
		comparisons["kda"] = ComparisonHigher
	} else {
		comparisons["kda"] = ComparisonLower
	}

	// Champion comparison - exact match only
	if guessed.PlayerMostPlayedChampion == target.PlayerMostPlayedChampion {
		comparisons["champion"] = ComparisonExact
	} else {
		comparisons["champion"] = ComparisonWrong
	}

	// Average kills comparison - exact, higher, or lower (with tolerance)
	killsTolerance := 0.1
	if abs(guessed.AvgKills-target.AvgKills) < killsTolerance {
		comparisons["avg_kills"] = ComparisonExact
	} else if guessed.AvgKills > target.AvgKills {
		comparisons["avg_kills"] = ComparisonHigher
	} else {
		comparisons["avg_kills"] = ComparisonLower
	}

	// Average deaths comparison - exact, higher, or lower (with tolerance)
	deathsTolerance := 0.1
	if abs(guessed.AvgDeaths-target.AvgDeaths) < deathsTolerance {
		comparisons["avg_deaths"] = ComparisonExact
	} else if guessed.AvgDeaths > target.AvgDeaths {
		comparisons["avg_deaths"] = ComparisonHigher
	} else {
		comparisons["avg_deaths"] = ComparisonLower
	}

	// Average assists comparison - exact, higher, or lower (with tolerance)
	assistsTolerance := 0.1
	if abs(guessed.AvgAssists-target.AvgAssists) < assistsTolerance {
		comparisons["avg_assists"] = ComparisonExact
	} else if guessed.AvgAssists > target.AvgAssists {
		comparisons["avg_assists"] = ComparisonHigher
	} else {
		comparisons["avg_assists"] = ComparisonLower
	}

	// Year of birth comparison - exact, higher, or lower
	// Note: Higher birth year = younger age, so we compare inversely for age logic
	if guessed.YearOfBirth == target.YearOfBirth {
		comparisons["year_of_birth"] = ComparisonExact
	} else if guessed.YearOfBirth > target.YearOfBirth {
		// Guessed player is younger (born later), so target is older (lower age arrow)
		comparisons["year_of_birth"] = ComparisonLower
	} else {
		// Guessed player is older (born earlier), so target is younger (higher age arrow)
		comparisons["year_of_birth"] = ComparisonHigher
	}

	// Last split result comparison - exact, higher, or lower (as ranking: lower number = better)
	guessedRank := parseRankingToInt(guessed.LastSplitResult)
	targetRank := parseRankingToInt(target.LastSplitResult)
	if guessedRank == targetRank {
		comparisons["last_split_result"] = ComparisonExact
	} else if guessedRank > targetRank {
		// Higher number = worse ranking, so target is better (higher)
		comparisons["last_split_result"] = ComparisonHigher
	} else {
		// Lower number = better ranking, so target is worse (lower)
		comparisons["last_split_result"] = ComparisonLower
	}

	// First split in league comparison - exact, higher, or lower (year: higher = more recent)
	if guessed.FirstSplitInLeague == target.FirstSplitInLeague {
		comparisons["first_split_in_league"] = ComparisonExact
	} else if guessed.FirstSplitInLeague > target.FirstSplitInLeague {
		// Guessed year is higher (more recent), target is earlier (lower)
		comparisons["first_split_in_league"] = ComparisonLower
	} else {
		// Guessed year is lower (earlier), target is more recent (higher)
		comparisons["first_split_in_league"] = ComparisonHigher
	}

	return comparisons
}

// parseRankingToInt converts ranking string to integer for comparison
func parseRankingToInt(ranking string) int {
	// Remove any non-digit characters and parse as int
	rankStr := ""
	for _, char := range ranking {
		if char >= '0' && char <= '9' {
			rankStr += string(char)
		}
	}

	if rankStr == "" {
		return 999 // Default high value for invalid rankings
	}

	rank := 0
	for _, char := range rankStr {
		rank = rank*10 + int(char-'0')
	}

	return rank
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// handleCorrectGuess processes a correct guess
func (gs *GameSession) handleCorrectGuess() {
	// Calculate current total elapsed time
	totalElapsed := gs.GetTotalElapsedTime()

	// Calculate wrong guesses for current player (total guesses - 1 for the correct guess)
	wrongGuesses := len(gs.Guesses) - 1
	if wrongGuesses < 0 {
		wrongGuesses = 0
	}

	// Award points for finding this player
	playerPoints := CalculatePlayerPoints(totalElapsed, wrongGuesses)
	gs.Score += playerPoints

	log.Printf("CORRECT GUESS in session %s! Player %d completed with %d wrong guesses. Points: %d (Total: %d, Time: %ds/%ds)",
		gs.SessionID, gs.CurrentPlayerIndex+1, wrongGuesses, playerPoints, gs.Score, totalElapsed, TotalGameTime)

	// Move to next player using the proper function
	if !gs.MoveToNextPlayer() {
		// All players completed
		gs.CompleteSession()
	}
}

// handleTimeLimit processes when time limit is reached for current player
func (gs *GameSession) handleTimeLimit() {
	log.Printf("Time limit reached in session %s for player %d/%d (2 minutes elapsed)",
		gs.SessionID, gs.CurrentPlayerIndex+1, len(gs.SelectedPlayers))

	// Move to next player or end session
	if !gs.MoveToNextPlayer() {
		gs.CompleteSession()
	}
}

// GetTimeRemaining returns the remaining time in seconds for the total game
func GetTimeRemaining(session *GameSession) int {
	if session == nil {
		return 0
	}

	elapsedSeconds := session.GetTotalElapsedTime()
	remainingSeconds := TotalGameTime - elapsedSeconds

	if remainingSeconds < 0 {
		return 0
	}

	return remainingSeconds
}
