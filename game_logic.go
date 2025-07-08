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
		return false
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
		completionBonus := 10000
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

	targetPlayer := session.GetCurrentPlayer()
	if targetPlayer == nil {
		return nil, fmt.Errorf("no current target player")
	}

	comparisons := comparePlayersDetailed(*guessedPlayer, *targetPlayer)
	isCorrect := guessedPlayer.PlayerUsername == targetPlayer.PlayerUsername

	guessResult := GuessResult{
		GuessedPlayer: *guessedPlayer,
		TargetPlayer:  *targetPlayer,
		Timestamp:     time.Now(),
		Comparisons:   comparisons,
		IsCorrect:     isCorrect,
	}

	session.Guesses = append(session.Guesses, guessResult)

	if isCorrect {
		session.handleCorrectGuess()
	} else {

		if session.IsGameOver() {

			session.handleTimeLimit()
		}
	}

	UpdateSession(session)

	return &guessResult, nil
}

func comparePlayersDetailed(guessed, target Player) map[string]ComparisonResult {
	comparisons := make(map[string]ComparisonResult)

	if guessed.PlayerTeam == target.PlayerTeam {
		comparisons["team"] = ComparisonExact
	} else if guessed.PlayerLeague == target.PlayerLeague {
		comparisons["team"] = ComparisonPartial
	} else {
		comparisons["team"] = ComparisonWrong
	}

	if guessed.PlayerLeague == target.PlayerLeague {
		comparisons["league"] = ComparisonExact
	} else {
		comparisons["league"] = ComparisonWrong
	}

	if guessed.PlayerRole == target.PlayerRole {
		comparisons["role"] = ComparisonExact
	} else {
		comparisons["role"] = ComparisonWrong
	}

	if guessed.PlayerCountry == target.PlayerCountry {
		comparisons["country"] = ComparisonExact
	} else if guessed.PlayerCountryContinent == target.PlayerCountryContinent {
		comparisons["country"] = ComparisonPartial
	} else {
		comparisons["country"] = ComparisonWrong
	}

	if guessed.PlayerAge == target.PlayerAge {
		comparisons["age"] = ComparisonExact
	} else if guessed.PlayerAge > target.PlayerAge {
		comparisons["age"] = ComparisonHigher
	} else {
		comparisons["age"] = ComparisonLower
	}

	if guessed.NumberOfClubs == target.NumberOfClubs {
		comparisons["clubs"] = ComparisonExact
	} else if guessed.NumberOfClubs > target.NumberOfClubs {
		comparisons["clubs"] = ComparisonHigher
	} else {
		comparisons["clubs"] = ComparisonLower
	}

	kdaTolerance := 0.01
	if abs(guessed.KDARatio-target.KDARatio) < kdaTolerance {
		comparisons["kda"] = ComparisonExact
	} else if guessed.KDARatio > target.KDARatio {
		comparisons["kda"] = ComparisonHigher
	} else {
		comparisons["kda"] = ComparisonLower
	}

	if guessed.PlayerMostPlayedChampion == target.PlayerMostPlayedChampion {
		comparisons["champion"] = ComparisonExact
	} else {
		comparisons["champion"] = ComparisonWrong
	}

	killsTolerance := 0.1
	if abs(guessed.AvgKills-target.AvgKills) < killsTolerance {
		comparisons["avg_kills"] = ComparisonExact
	} else if guessed.AvgKills > target.AvgKills {
		comparisons["avg_kills"] = ComparisonHigher
	} else {
		comparisons["avg_kills"] = ComparisonLower
	}

	deathsTolerance := 0.1
	if abs(guessed.AvgDeaths-target.AvgDeaths) < deathsTolerance {
		comparisons["avg_deaths"] = ComparisonExact
	} else if guessed.AvgDeaths > target.AvgDeaths {
		comparisons["avg_deaths"] = ComparisonHigher
	} else {
		comparisons["avg_deaths"] = ComparisonLower
	}

	assistsTolerance := 0.1
	if abs(guessed.AvgAssists-target.AvgAssists) < assistsTolerance {
		comparisons["avg_assists"] = ComparisonExact
	} else if guessed.AvgAssists > target.AvgAssists {
		comparisons["avg_assists"] = ComparisonHigher
	} else {
		comparisons["avg_assists"] = ComparisonLower
	}

	if guessed.YearOfBirth == target.YearOfBirth {
		comparisons["year_of_birth"] = ComparisonExact
	} else if guessed.YearOfBirth > target.YearOfBirth {

		comparisons["year_of_birth"] = ComparisonLower
	} else {

		comparisons["year_of_birth"] = ComparisonHigher
	}

	guessedRank := parseRankingToInt(guessed.LastSplitResult)
	targetRank := parseRankingToInt(target.LastSplitResult)
	if guessedRank == targetRank {
		comparisons["last_split_result"] = ComparisonExact
	} else if guessedRank > targetRank {

		comparisons["last_split_result"] = ComparisonHigher
	} else {

		comparisons["last_split_result"] = ComparisonLower
	}

	if guessed.FirstSplitInLeague == target.FirstSplitInLeague {
		comparisons["first_split_in_league"] = ComparisonExact
	} else if guessed.FirstSplitInLeague > target.FirstSplitInLeague {

		comparisons["first_split_in_league"] = ComparisonLower
	} else {

		comparisons["first_split_in_league"] = ComparisonHigher
	}

	return comparisons
}

func parseRankingToInt(ranking string) int {

	rankStr := ""
	for _, char := range ranking {
		if char >= '0' && char <= '9' {
			rankStr += string(char)
		}
	}

	if rankStr == "" {
		return 999
	}

	rank := 0
	for _, char := range rankStr {
		rank = rank*10 + int(char-'0')
	}

	return rank
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func (gs *GameSession) handleCorrectGuess() {

	totalElapsed := gs.GetTotalElapsedTime()

	wrongGuesses := len(gs.Guesses) - 1
	if wrongGuesses < 0 {
		wrongGuesses = 0
	}

	playerPoints := CalculatePlayerPoints(totalElapsed, wrongGuesses)
	gs.Score += playerPoints

	if !gs.MoveToNextPlayer() {

		gs.CompleteSession()
	}
}

func (gs *GameSession) handleTimeLimit() {
	log.Printf("Time limit reached in session %s for player %d/%d (2 minutes elapsed)",
		gs.SessionID, gs.CurrentPlayerIndex+1, len(gs.SelectedPlayers))

	if !gs.MoveToNextPlayer() {
		gs.CompleteSession()
	}
}

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
