package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// InitDatabase initializes the SQLite database connection and creates tables
func InitDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "./prodle.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Create tables
	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// createTables creates the necessary database tables
func createTables() error {
	// Create leaderboard table
	leaderboardQuery := `
	CREATE TABLE IF NOT EXISTS leaderboard (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		score INTEGER NOT NULL,
		date DATETIME NOT NULL,
		duration INTEGER NOT NULL,
		guess_count INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(leaderboardQuery); err != nil {
		return fmt.Errorf("failed to create leaderboard table: %v", err)
	}

	// Create index on score for faster leaderboard queries
	indexQuery := `
	CREATE INDEX IF NOT EXISTS idx_leaderboard_score 
	ON leaderboard(score DESC, duration ASC);`

	if _, err := db.Exec(indexQuery); err != nil {
		return fmt.Errorf("failed to create leaderboard index: %v", err)
	}

	return nil
}

// GetDatabase returns the database connection
func GetDatabase() *sql.DB {
	return db
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// AddLeaderboardEntry adds a new entry to the leaderboard
func AddLeaderboardEntry(entry LeaderboardEntry) error {
	query := `
	INSERT INTO leaderboard (username, score, date, duration, guess_count)
	VALUES (?, ?, ?, ?, ?)`

	_, err := db.Exec(query, entry.Username, entry.Score, entry.Date, entry.Duration, entry.GuessCount)
	if err != nil {
		return fmt.Errorf("failed to add leaderboard entry: %v", err)
	}

	return nil
}

// AddToLeaderboard adds a new entry to the leaderboard with validation
func AddToLeaderboard(username string, score int) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if score < 0 {
		score = 0 // Ensure score is not negative
	}

	entry := LeaderboardEntry{
		Username:   SanitizeInput(username),
		Score:      score,
		Date:       time.Now(),
		Duration:   0, // Will be calculated from session data if available
		GuessCount: 0, // Will be calculated from session data if available
	}

	return AddLeaderboardEntry(entry)
}

// AddToLeaderboardFromSession adds a leaderboard entry from a completed game session
func AddToLeaderboardFromSession(username string, session *GameSession) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	// Calculate total duration and guess count
	var totalDuration int
	var totalGuesses int

	if session.CompletionTime != nil {
		totalDuration = int(session.CompletionTime.Sub(session.StartTime).Seconds())
	} else {
		totalDuration = int(time.Since(session.StartTime).Seconds())
	}

	// Count total guesses across all players attempted
	totalGuesses = len(session.Guesses)

	entry := LeaderboardEntry{
		Username:   SanitizeInput(username),
		Score:      session.Score,
		Date:       time.Now(),
		Duration:   totalDuration,
		GuessCount: totalGuesses,
	}

	return AddLeaderboardEntry(entry)
}

// GetTop10Scores retrieves the top 10 leaderboard entries ordered by score DESC
func GetTop10Scores() ([]LeaderboardEntry, error) {
	return GetLeaderboard(10)
}

// GetLeaderboard retrieves the top leaderboard entries
func GetLeaderboard(limit int) ([]LeaderboardEntry, error) {
	query := `
	SELECT username, score, date, duration, guess_count
	FROM leaderboard
	ORDER BY score DESC, duration ASC
	LIMIT ?`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %v", err)
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		err := rows.Scan(
			&entry.Username,
			&entry.Score,
			&entry.Date,
			&entry.Duration,
			&entry.GuessCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %v", err)
		}
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating leaderboard rows: %v", err)
	}

	return entries, nil
}

// GetUserBestScore retrieves the best score for a specific user
func GetUserBestScore(username string) (*LeaderboardEntry, error) {
	query := `
	SELECT username, score, date, duration, guess_count
	FROM leaderboard
	WHERE username = ?
	ORDER BY score DESC, duration ASC
	LIMIT 1`

	row := db.QueryRow(query, username)

	var entry LeaderboardEntry
	err := row.Scan(
		&entry.Username,
		&entry.Score,
		&entry.Date,
		&entry.Duration,
		&entry.GuessCount,
	)

	if err == sql.ErrNoRows {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user best score: %v", err)
	}

	return &entry, nil
}

// GetLeaderboardStats returns basic statistics about the leaderboard
func GetLeaderboardStats() (map[string]interface{}, error) {
	query := `
	SELECT 
		COUNT(*) as total_games,
		AVG(score) as avg_score,
		MAX(score) as highest_score,
		AVG(duration) as avg_duration,
		AVG(guess_count) as avg_guesses
	FROM leaderboard`

	row := db.QueryRow(query)

	var stats struct {
		TotalGames   int
		AvgScore     float64
		HighestScore int
		AvgDuration  float64
		AvgGuesses   float64
	}

	err := row.Scan(
		&stats.TotalGames,
		&stats.AvgScore,
		&stats.HighestScore,
		&stats.AvgDuration,
		&stats.AvgGuesses,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard stats: %v", err)
	}

	result := map[string]interface{}{
		"total_games":   stats.TotalGames,
		"avg_score":     stats.AvgScore,
		"highest_score": stats.HighestScore,
		"avg_duration":  stats.AvgDuration,
		"avg_guesses":   stats.AvgGuesses,
	}

	return result, nil
}

// FormattedLeaderboardEntry represents a leaderboard entry with formatted display data
type FormattedLeaderboardEntry struct {
	Rank              int    `json:"rank"`
	Username          string `json:"username"`
	Score             int    `json:"score"`
	FormattedDate     string `json:"formatted_date"`
	FormattedDuration string `json:"formatted_duration"`
	GuessCount        int    `json:"guess_count"`
}

// GetFormattedLeaderboard returns leaderboard entries formatted for display
func GetFormattedLeaderboard(limit int) ([]FormattedLeaderboardEntry, error) {
	entries, err := GetLeaderboard(limit)
	if err != nil {
		return nil, err
	}

	formatted := make([]FormattedLeaderboardEntry, len(entries))
	for i, entry := range entries {
		formatted[i] = FormattedLeaderboardEntry{
			Rank:              i + 1,
			Username:          entry.Username,
			Score:             entry.Score,
			FormattedDate:     entry.Date.Format("Jan 2, 2006"),
			FormattedDuration: FormatDuration(entry.Duration),
			GuessCount:        entry.GuessCount,
		}
	}

	return formatted, nil
}

// GetFormattedTop10 returns the top 10 scores formatted for display
func GetFormattedTop10() ([]FormattedLeaderboardEntry, error) {
	return GetFormattedLeaderboard(10)
}
