package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

// In-memory session storage
var (
	activeSessions = make(map[string]*GameSession)
	sessionMutex   sync.RWMutex
)

const (
	MaxGuesses        = 6
	SessionTimeout    = 24 * time.Hour // Sessions expire after 24 hours
	CleanupInterval   = 1 * time.Hour  // Run cleanup every hour
	PlayersPerSession = 20
)

// SessionManager handles session lifecycle
type SessionManager struct {
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// NewSessionManager creates a new session manager with automatic cleanup
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		cleanupTicker: time.NewTicker(CleanupInterval),
		stopCleanup:   make(chan bool),
	}

	// Start background cleanup goroutine
	go sm.runCleanup()

	return sm
}

// runCleanup runs in background to clean up expired sessions
func (sm *SessionManager) runCleanup() {
	for {
		select {
		case <-sm.cleanupTicker.C:
			sm.cleanupExpiredSessions()
		case <-sm.stopCleanup:
			sm.cleanupTicker.Stop()
			return
		}
	}
}

// Stop stops the session manager cleanup process
func (sm *SessionManager) Stop() {
	sm.stopCleanup <- true
}

// cleanupExpiredSessions removes sessions that have expired
func (sm *SessionManager) cleanupExpiredSessions() {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for sessionID, session := range activeSessions {
		if now.Sub(session.StartTime) > SessionTimeout {
			delete(activeSessions, sessionID)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		log.Printf("Cleaned up %d expired sessions", expiredCount)
	}
}

// generateSessionID creates a unique session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateNewSession creates a new game session with 20 random players
func CreateNewSession() (*GameSession, error) {
	// Generate unique session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %v", err)
	}

	// Get random players for this session
	players, err := GetRandomPlayers(PlayersPerSession)
	if err != nil {
		return nil, fmt.Errorf("failed to get random players: %v", err)
	}

	// Create new session
	now := time.Now()
	session := &GameSession{
		SessionID:              sessionID,
		SelectedPlayers:        players,
		CurrentPlayerIndex:     0,
		Score:                  0,
		StartTime:              now,
		CurrentPlayerStartTime: now,
		Guesses:                make([]GuessResult, 0),
		IsCompleted:            false,
		CompletionTime:         nil,
	}

	// Store session in memory
	sessionMutex.Lock()
	activeSessions[sessionID] = session
	sessionMutex.Unlock()

	log.Printf("Created new session %s with %d players", sessionID, len(players))
	return session, nil
}

// GetSession retrieves a session by ID
func GetSession(sessionID string) (*GameSession, bool) {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()

	session, exists := activeSessions[sessionID]
	return session, exists
}

// UpdateSession updates a session in storage
func UpdateSession(session *GameSession) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	activeSessions[session.SessionID] = session
}

// DeleteSession removes a session from storage
func DeleteSession(sessionID string) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	delete(activeSessions, sessionID)
}

// GetCurrentPlayer returns the current target player for guessing
func (gs *GameSession) GetCurrentPlayer() *Player {
	if gs.CurrentPlayerIndex >= 0 && gs.CurrentPlayerIndex < len(gs.SelectedPlayers) {
		return &gs.SelectedPlayers[gs.CurrentPlayerIndex]
	}
	return nil
}

// GetCurrentPlayerElapsedTime returns the elapsed time for the current player in seconds
func (gs *GameSession) GetCurrentPlayerElapsedTime() int {
	return int(time.Since(gs.CurrentPlayerStartTime).Seconds())
}

// GetCurrentPlayerScore returns the potential score for the current player based on elapsed time and wrong guesses
func (gs *GameSession) GetCurrentPlayerScore() int {
	elapsedSeconds := gs.GetCurrentPlayerElapsedTime()
	wrongGuesses := len(gs.Guesses) // All current guesses are wrong since we haven't found the correct answer yet
	return CalculatePlayerScore(elapsedSeconds, wrongGuesses)
}

// CheckCorrectGuess checks if the guessed player is correct
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

// MoveToNextPlayer moves to the next player in the session
func (gs *GameSession) MoveToNextPlayer() bool {
	gs.CurrentPlayerIndex++

	// Check if all players completed
	if gs.CurrentPlayerIndex >= len(gs.SelectedPlayers) {
		return false // Game completed
	}

	// Reset for next player
	gs.Guesses = make([]GuessResult, 0)
	gs.CurrentPlayerStartTime = time.Now()

	log.Printf("Session %s moved to player %d/%d",
		gs.SessionID, gs.CurrentPlayerIndex+1, len(gs.SelectedPlayers))

	return true // More players remaining
}

// IsGameOver checks if the game should end (2 minutes elapsed for current player)
func (gs *GameSession) IsGameOver() bool {
	if gs.IsCompleted {
		return true
	}

	// Check if 2 minutes (120 seconds) have elapsed for current player
	elapsedSeconds := gs.GetCurrentPlayerElapsedTime()
	if elapsedSeconds >= 120 {
		return true
	}

	// Check if max guesses reached
	if len(gs.Guesses) >= MaxGuesses {
		return true
	}

	// Check if all players completed
	if gs.CurrentPlayerIndex >= len(gs.SelectedPlayers) {
		return true
	}

	return false
}

// CalculateFinalScore calculates and sets the final score when game ends
func (gs *GameSession) CalculateFinalScore() int {
	if !gs.IsCompleted {
		// If game is ending due to timeout or max guesses, don't add score for current player
		log.Printf("Game ending without completion for session %s at player %d/%d",
			gs.SessionID, gs.CurrentPlayerIndex+1, len(gs.SelectedPlayers))
	}

	// Final score is already calculated incrementally during the game
	// Add any completion bonuses here if needed
	completedPlayers := gs.CurrentPlayerIndex
	if gs.IsCompleted && completedPlayers > len(gs.SelectedPlayers) {
		completedPlayers = len(gs.SelectedPlayers)
	}

	// Bonus for completing all players
	if completedPlayers == len(gs.SelectedPlayers) {
		completionBonus := 1000
		gs.Score += completionBonus
		log.Printf("Session %s completed all players! Bonus: %d points", gs.SessionID, completionBonus)
	}

	return gs.Score
}

// completeSession marks the session as completed
func (gs *GameSession) completeSession() {
	gs.IsCompleted = true
	now := time.Now()
	gs.CompletionTime = &now

	// Calculate final score with bonuses
	finalScore := gs.CalculateFinalScore()

	duration := int(time.Since(gs.StartTime).Seconds())
	completedPlayers := gs.CurrentPlayerIndex
	if completedPlayers > len(gs.SelectedPlayers) {
		completedPlayers = len(gs.SelectedPlayers)
	}

	log.Printf("Session %s completed. Players: %d/%d, Final Score: %d, Duration: %ds",
		gs.SessionID, completedPlayers, len(gs.SelectedPlayers), finalScore, duration)
}

// ValidateGuess validates a player guess against the target and returns detailed comparison results
func ValidateGuess(session *GameSession, guessedPlayerName string) (*GuessResult, error) {
	// Check if session is valid and not completed
	if session == nil {
		return nil, fmt.Errorf("session is nil")
	}

	if session.IsCompleted {
		return nil, fmt.Errorf("session already completed")
	}

	// Check if max guesses reached for current player
	currentPlayerGuesses := len(session.Guesses)
	if currentPlayerGuesses >= MaxGuesses {
		return nil, fmt.Errorf("maximum guesses reached for current player")
	}

	// Validate and sanitize input
	guessedPlayerName = SanitizeInput(guessedPlayerName)
	if valid, errMsg := ValidatePlayerGuess(guessedPlayerName); !valid {
		return nil, fmt.Errorf("invalid guess: %s", errMsg)
	}

	// Get the guessed player
	guessedPlayer, exists := GetPlayerByName(guessedPlayerName)
	if !exists {
		return nil, fmt.Errorf("player not found: %s", guessedPlayerName)
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
		// Check if game should end after this guess
		if session.IsGameOver() {
			if len(session.Guesses) >= MaxGuesses {
				session.handleMaxGuessesReached()
			} else {
				// Time limit reached
				session.handleTimeLimit()
			}
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

	return comparisons
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
	// Calculate elapsed time for this player
	elapsedSeconds := int(time.Since(gs.CurrentPlayerStartTime).Seconds())

	// Calculate wrong guesses (total guesses - 1 for the correct guess)
	wrongGuesses := len(gs.Guesses) - 1
	if wrongGuesses < 0 {
		wrongGuesses = 0
	}

	// Calculate score for this player using new scoring system
	playerScore := CalculatePlayerScore(elapsedSeconds, wrongGuesses)
	gs.Score += playerScore

	// Move to next player
	gs.CurrentPlayerIndex++

	log.Printf("Correct guess in session %s! Player %d completed in %ds with %d wrong guesses. Score: %d (Total: %d)",
		gs.SessionID, gs.CurrentPlayerIndex, elapsedSeconds, wrongGuesses, playerScore, gs.Score)

	// Check if all players completed
	if gs.CurrentPlayerIndex >= len(gs.SelectedPlayers) {
		gs.completeSession()
	} else {
		// Reset guesses and start time for next player
		gs.Guesses = make([]GuessResult, 0)
		gs.CurrentPlayerStartTime = time.Now() // Reset timer for next player
	}
}

// handleMaxGuessesReached processes when max guesses are reached without correct answer
func (gs *GameSession) handleMaxGuessesReached() {
	log.Printf("Max guesses reached in session %s for player %d/%d",
		gs.SessionID, gs.CurrentPlayerIndex+1, len(gs.SelectedPlayers))

	// Move to next player or end session
	if !gs.MoveToNextPlayer() {
		gs.completeSession()
	}
}

// handleTimeLimit processes when time limit is reached for current player
func (gs *GameSession) handleTimeLimit() {
	log.Printf("Time limit reached in session %s for player %d/%d (2 minutes elapsed)",
		gs.SessionID, gs.CurrentPlayerIndex+1, len(gs.SelectedPlayers))

	// Move to next player or end session
	if !gs.MoveToNextPlayer() {
		gs.completeSession()
	}
}

// GetSessionStats returns statistics about all active sessions
func GetSessionStats() map[string]interface{} {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()

	totalSessions := len(activeSessions)
	completedSessions := 0
	activePlaying := 0

	for _, session := range activeSessions {
		if session.IsCompleted {
			completedSessions++
		} else {
			activePlaying++
		}
	}

	return map[string]interface{}{
		"total_sessions":     totalSessions,
		"completed_sessions": completedSessions,
		"active_playing":     activePlaying,
	}
}

// GetTimeRemaining returns the remaining time in seconds for the current player
func GetTimeRemaining(session *GameSession) int {
	if session == nil {
		return 0
	}

	elapsedSeconds := int(time.Since(session.CurrentPlayerStartTime).Seconds())
	remainingSeconds := 120 - elapsedSeconds // 2 minutes = 120 seconds

	if remainingSeconds < 0 {
		return 0
	}

	return remainingSeconds
}

// GetCurrentPlayerGuesses returns the number of guesses made for the current player
func GetCurrentPlayerGuesses(session *GameSession) int {
	if session == nil {
		return 0
	}

	return len(session.Guesses)
}

// IsGameOver is a wrapper function for the session method
func IsGameOver(session *GameSession) bool {
	if session == nil {
		return true
	}

	return session.IsGameOver()
}

// CheckCorrectGuess checks if the guessed player name matches the current target player
func CheckCorrectGuess(session *GameSession, playerName string) bool {
	if session == nil {
		return false
	}

	targetPlayer := session.GetCurrentPlayer()
	if targetPlayer == nil {
		return false
	}

	guessedPlayer, exists := GetPlayerByName(playerName)
	if !exists {
		return false
	}

	return guessedPlayer.PlayerUsername == targetPlayer.PlayerUsername
}
