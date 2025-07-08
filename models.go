package main

import (
	"time"
)

// TournamentWin represents a tournament victory
type TournamentWin struct {
	Tournament string `json:"tournament"`
	Date       string `json:"date"`
	Team       string `json:"team"`
}

// Player represents a professional League of Legends player
type Player struct {
	ID                    string          `json:"ID"`
	Team                  string          `json:"Team"`
	League                string          `json:"League"`
	YearOfBirth           int             `json:"YearOfBirth"`
	Role                  string          `json:"Role"`
	Nationality           string          `json:"Nationality"`
	Continent             string          `json:"Continent"`
	LastSplitResult       string          `json:"LastSplitResult"`
	FirstSplitInLeague    int             `json:"FirstSplitInLeague"`
	TeamsPlayed           []string        `json:"TeamsPlayed"`
	PrimaryTournamentWins []TournamentWin `json:"PrimaryTournamentWins"`

	// Computed fields for game logic
	PlayerUsername           string  // Will be set to ID for compatibility
	PlayerName               string  // Will be set to ID for compatibility
	PlayerMediaURL           string  // Legacy field, can be empty
	PlayerTeam               string  // Will be set to Team for compatibility
	PlayerTeamMediaURL       string  // Legacy field, can be empty
	PlayerLeague             string  // Will be set to League for compatibility
	NumberOfClubs            int     // Will be set to len(TeamsPlayed) for compatibility
	PlayerCountry            string  // Will be set to Nationality for compatibility
	PlayerCountryContinent   string  // Will be set to Continent for compatibility
	PlayerRole               string  // Will be set to Role for compatibility
	PlayerMostPlayedChampion string  // Legacy field, can be empty
	PlayerAge                int     // Will be computed from YearOfBirth
	AvgKills                 float64 // Legacy field, can be 0
	AvgDeaths                float64 // Legacy field, can be 0
	AvgAssists               float64 // Legacy field, can be 0
	KDARatio                 float64 // Legacy field, can be 0
	GamesPlayed              int     // Legacy field, can be 0
}

// ComparisonResult represents the result of comparing a guess attribute with the target
type ComparisonResult string

const (
	ComparisonExact   ComparisonResult = "exact"   // Exact match
	ComparisonHigher  ComparisonResult = "higher"  // Guess is higher than target
	ComparisonLower   ComparisonResult = "lower"   // Guess is lower than target
	ComparisonPartial ComparisonResult = "partial" // Partial match (e.g., same continent but different country)
	ComparisonWrong   ComparisonResult = "wrong"   // No match
)

// GuessResult contains the result of a player guess with comparison results
type GuessResult struct {
	GuessedPlayer Player                      `json:"guessed_player"`
	TargetPlayer  Player                      `json:"-"` // Don't send to client - security
	Timestamp     time.Time                   `json:"timestamp"`
	Comparisons   map[string]ComparisonResult `json:"comparisons"`
	IsCorrect     bool                        `json:"is_correct"`
}

// GameSession represents an individual game session
type GameSession struct {
	SessionID          string        `json:"session_id"`
	Difficulty         string        `json:"difficulty"`       // Difficulty level: facile, moyen, difficile
	SelectedPlayers    []Player      `json:"selected_players"` // 20 players for the session
	CurrentPlayerIndex int           `json:"current_player_index"`
	Score              int           `json:"score"`
	StartTime          time.Time     `json:"start_time"` // Total game start time
	Guesses            []GuessResult `json:"guesses"`
	IsCompleted        bool          `json:"is_completed"`
	CompletionTime     *time.Time    `json:"completion_time,omitempty"`
}

// LeaderboardEntry represents a single entry in the leaderboard
type LeaderboardEntry struct {
	Username   string    `json:"username"`
	Score      int       `json:"score"`
	Date       time.Time `json:"date"`
	Duration   int       `json:"duration"` // Duration in seconds
	GuessCount int       `json:"guess_count"`
}
