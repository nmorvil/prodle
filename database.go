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
	// Create separate leaderboard tables for each difficulty
	difficulties := []string{"facile", "moyen", "difficile"}

	for _, difficulty := range difficulties {
		leaderboardQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS leaderboard_%s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			score INTEGER NOT NULL,
			date DATETIME NOT NULL,
			duration INTEGER NOT NULL,
			guess_count INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`, difficulty)

		if _, err := db.Exec(leaderboardQuery); err != nil {
			return fmt.Errorf("failed to create leaderboard_%s table: %v", difficulty, err)
		}
	}

	// Create legacy leaderboard table for backwards compatibility
	legacyLeaderboardQuery := `
	CREATE TABLE IF NOT EXISTS leaderboard (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		score INTEGER NOT NULL,
		date DATETIME NOT NULL,
		duration INTEGER NOT NULL,
		guess_count INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(legacyLeaderboardQuery); err != nil {
		return fmt.Errorf("failed to create leaderboard table: %v", err)
	}

	// Create indexes on score for faster leaderboard queries
	for _, difficulty := range difficulties {
		indexQuery := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS idx_leaderboard_%s_score 
		ON leaderboard_%s(score DESC, duration ASC);`, difficulty, difficulty)

		if _, err := db.Exec(indexQuery); err != nil {
			return fmt.Errorf("failed to create leaderboard_%s index: %v", difficulty, err)
		}
	}

	// Create legacy index
	legacyIndexQuery := `
	CREATE INDEX IF NOT EXISTS idx_leaderboard_score 
	ON leaderboard(score DESC, duration ASC);`

	if _, err := db.Exec(legacyIndexQuery); err != nil {
		return fmt.Errorf("failed to create legacy leaderboard index: %v", err)
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

// AddLeaderboardEntryByDifficulty adds a new entry to the difficulty-specific leaderboard
func AddLeaderboardEntryByDifficulty(entry LeaderboardEntry, difficulty string) error {
	// Validate difficulty
	validDifficulties := map[string]bool{
		"facile":    true,
		"moyen":     true,
		"difficile": true,
	}

	if !validDifficulties[difficulty] {
		return fmt.Errorf("invalid difficulty: %s", difficulty)
	}

	query := fmt.Sprintf(`
	INSERT INTO leaderboard_%s (username, score, date, duration, guess_count)
	VALUES (?, ?, ?, ?, ?)`, difficulty)

	_, err := db.Exec(query, entry.Username, entry.Score, entry.Date, entry.Duration, entry.GuessCount)
	if err != nil {
		return fmt.Errorf("failed to add leaderboard_%s entry: %v", difficulty, err)
	}

	return nil
}

// AddToLeaderboardByDifficulty adds a new entry to the difficulty-specific leaderboard with validation
func AddToLeaderboardByDifficulty(username string, score int, difficulty string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if score < 0 {
		return fmt.Errorf("score cannot be negative")
	}

	entry := LeaderboardEntry{
		Username:   SanitizeInput(username),
		Score:      score,
		Date:       time.Now(),
		Duration:   0, // Will be updated if needed
		GuessCount: 0, // Will be updated if needed
	}

	return AddLeaderboardEntryByDifficulty(entry, difficulty)
}

// SubmitScoreByDifficulty adds a score to the difficulty-specific leaderboard from a game session
func SubmitScoreByDifficulty(username string, session *GameSession, difficulty string) error {
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

	return AddLeaderboardEntryByDifficulty(entry, difficulty)
}

// GetLeaderboardByDifficulty retrieves the top leaderboard entries for a specific difficulty
func GetLeaderboardByDifficulty(limit int, difficulty string) ([]LeaderboardEntry, error) {
	// Validate difficulty
	validDifficulties := map[string]bool{
		"facile":    true,
		"moyen":     true,
		"difficile": true,
	}

	if !validDifficulties[difficulty] {
		return nil, fmt.Errorf("invalid difficulty: %s", difficulty)
	}

	query := fmt.Sprintf(`
	SELECT username, score, date, duration, guess_count
	FROM leaderboard_%s
	ORDER BY score DESC, duration ASC
	LIMIT ?`, difficulty)

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard_%s: %v", difficulty, err)
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
			return nil, fmt.Errorf("failed to scan leaderboard_%s row: %v", difficulty, err)
		}
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating leaderboard_%s rows: %v", difficulty, err)
	}

	return entries, nil
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

// GetFormattedLeaderboardByDifficulty returns leaderboard entries formatted for display for a specific difficulty
func GetFormattedLeaderboardByDifficulty(limit int, difficulty string) ([]FormattedLeaderboardEntry, error) {
	entries, err := GetLeaderboardByDifficulty(limit, difficulty)
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

// GetPlayerRankByDifficulty calculates the rank of a score in the leaderboard for a specific difficulty
func GetPlayerRankByDifficulty(score int, duration int, difficulty string) (int, error) {
	// Validate difficulty
	validDifficulties := map[string]bool{
		"facile":    true,
		"moyen":     true,
		"difficile": true,
	}

	if !validDifficulties[difficulty] {
		return 0, fmt.Errorf("invalid difficulty: %s", difficulty)
	}

	// Count how many scores are better than the given score
	// Better means: higher score, or same score with lower duration
	query := fmt.Sprintf(`
	SELECT COUNT(*) + 1 as rank
	FROM leaderboard_%s
	WHERE score > ? OR (score = ? AND duration < ?)`, difficulty)

	var rank int
	err := db.QueryRow(query, score, score, duration).Scan(&rank)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate rank for difficulty %s: %v", difficulty, err)
	}

	return rank, nil
}
