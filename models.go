package main

import (
	"time"
)

type TournamentWin struct {
	Tournament string `json:"tournament"`
	Date       string `json:"date"`
	Team       string `json:"team"`
}

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

	PlayerUsername           string
	PlayerName               string
	PlayerMediaURL           string
	PlayerTeam               string
	PlayerTeamMediaURL       string
	PlayerLeague             string
	NumberOfClubs            int
	PlayerCountry            string
	PlayerCountryContinent   string
	PlayerRole               string
	PlayerMostPlayedChampion string
	PlayerAge                int
	AvgKills                 float64
	AvgDeaths                float64
	AvgAssists               float64
	KDARatio                 float64
	GamesPlayed              int
}

type ComparisonResult string

const (
	ComparisonExact   ComparisonResult = "exact"
	ComparisonHigher  ComparisonResult = "higher"
	ComparisonLower   ComparisonResult = "lower"
	ComparisonPartial ComparisonResult = "partial"
	ComparisonWrong   ComparisonResult = "wrong"
)

type GuessResult struct {
	GuessedPlayer Player                      `json:"guessed_player"`
	TargetPlayer  Player                      `json:"-"`
	Timestamp     time.Time                   `json:"timestamp"`
	Comparisons   map[string]ComparisonResult `json:"comparisons"`
	IsCorrect     bool                        `json:"is_correct"`
}

type GameSession struct {
	SessionID          string        `json:"session_id"`
	Difficulty         string        `json:"difficulty"`
	SelectedPlayers    []Player      `json:"selected_players"`
	CurrentPlayerIndex int           `json:"current_player_index"`
	Score              int           `json:"score"`
	StartTime          time.Time     `json:"start_time"`
	Guesses            []GuessResult `json:"guesses"`
	IsCompleted        bool          `json:"is_completed"`
	CompletionTime     *time.Time    `json:"completion_time,omitempty"`
}

type LeaderboardEntry struct {
	Username   string    `json:"username"`
	Score      int       `json:"score"`
	Date       time.Time `json:"date"`
	Duration   int       `json:"duration"`
	GuessCount int       `json:"guess_count"`
}
