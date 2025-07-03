package main

import (
	"time"
)

// Player represents a professional League of Legends player
type Player struct {
	PlayerUsername           string  `json:"player_username"`
	PlayerName               string  `json:"player_name"`
	PlayerMediaURL           string  `json:"player_media_url"`
	PlayerTeam               string  `json:"player_team"`
	PlayerTeamMediaURL       string  `json:"player_team_media_url"`
	PlayerLeague             string  `json:"player_league"`
	NumberOfClubs            int     `json:"number_of_clubs"`
	PlayerCountry            string  `json:"player_country"`
	PlayerCountryContinent   string  `json:"player_country_continent"`
	PlayerRole               string  `json:"player_role"`
	PlayerMostPlayedChampion string  `json:"player_most_played_champion"`
	PlayerAge                int     `json:"player_age"`
	AvgKills                 float64 `json:"avg_kills"`
	AvgDeaths                float64 `json:"avg_deaths"`
	AvgAssists               float64 `json:"avg_assists"`
	KDARatio                 float64 `json:"kda_ratio"`
	GamesPlayed              int     `json:"games_played"`
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
