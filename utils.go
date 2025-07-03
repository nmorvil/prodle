package main

import (
	"fmt"
	"strings"
)

// GetChampionImg generates the champion image URL from the champion name
// Translates the Python function to Go with the same logic
func GetChampionImg(champion string) string {
	// Remove apostrophes, spaces, and periods
	curated := strings.ReplaceAll(champion, "'", "")
	curated = strings.ReplaceAll(curated, " ", "")
	curated = strings.ReplaceAll(curated, ".", "")

	// Handle special cases (equivalent to match/case in Python)
	switch curated {
	case "KaiSa":
		curated = "Kaisa"
	case "Wukong":
		curated = "MonkeyKing"
	case "RenataGlasc":
		curated = "Renata"
	}

	// Return the formatted URL
	return fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/img/champion/centered/%s_0.jpg", curated)
}

// Additional utility functions for the game

// FormatDuration converts seconds to a human-readable duration string
func FormatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	}

	minutes := seconds / 60
	remainingSeconds := seconds % 60

	if minutes < 60 {
		if remainingSeconds == 0 {
			return fmt.Sprintf("%d minutes", minutes)
		}
		return fmt.Sprintf("%d minutes %d seconds", minutes, remainingSeconds)
	}

	hours := minutes / 60
	remainingMinutes := minutes % 60

	if remainingMinutes == 0 && remainingSeconds == 0 {
		return fmt.Sprintf("%d hours", hours)
	} else if remainingSeconds == 0 {
		return fmt.Sprintf("%d hours %d minutes", hours, remainingMinutes)
	}

	return fmt.Sprintf("%d hours %d minutes %d seconds", hours, remainingMinutes, remainingSeconds)
}

// CalculatePlayerPoints calculates points for finding a single player
// More generous scoring with higher base points and smaller penalties
func CalculatePlayerPoints(totalElapsedSeconds, wrongGuesses int) int {
	// Much higher base points for finding a player
	basePoints := 5000

	// Reduce base points as total game time progresses (but less severely)
	timeProgress := float64(totalElapsedSeconds) / float64(TotalGameTime)
	if timeProgress > 1.0 {
		timeProgress = 1.0
	}

	// Start with 5000 points, decrease to 1500 points as time progresses (less steep decline)
	points := int(float64(basePoints) * (1.0 - 0.7*timeProgress))

	// Smaller penalty for wrong guesses (-100 points per wrong guess)
	penalty := wrongGuesses * 100
	points -= penalty

	// Higher minimum points per player found
	if points < 300 {
		points = 300
	}

	return points
}

// CalculateGameScore calculates total game score (for display during game)
func CalculateGameScore(totalElapsedSeconds, totalWrongGuesses, playersFound int) int {
	// This is mainly for display - actual scoring happens per player
	baseScore := playersFound * 3000        // Higher base points per player found
	timePenalty := totalElapsedSeconds * 10 // Modest time penalty
	guessPenalty := totalWrongGuesses * 50  // Smaller penalty for wrong guesses

	score := baseScore - timePenalty - guessPenalty
	if score < 0 {
		score = 0
	}

	return score
}

// Legacy function - keep for compatibility but update to use new system
func CalculatePoints(elapsedSeconds int) int {
	// For backwards compatibility, treat as if it's a single player game
	return CalculatePlayerPoints(elapsedSeconds, 0)
}

// CalculatePlayerScore calculates the final score for a player including wrong guess penalties
func CalculatePlayerScore(elapsedSeconds int, wrongGuesses int) int {
	basePoints := CalculatePoints(elapsedSeconds)

	// Apply penalty for wrong guesses (-100 per wrong guess)
	penalty := wrongGuesses * 100

	finalScore := basePoints - penalty

	// Ensure score doesn't go negative
	if finalScore < 0 {
		finalScore = 0
	}

	return finalScore
}

// ValidatePlayerGuess checks if a player guess is valid
func ValidatePlayerGuess(guess string) (bool, string) {
	guess = strings.TrimSpace(guess)

	if guess == "" {
		return false, "Player name cannot be empty"
	}

	if len(guess) < 2 {
		return false, "Player name must be at least 2 characters long"
	}

	if len(guess) > 50 {
		return false, "Player name is too long"
	}

	return true, ""
}

// SanitizeInput sanitizes user input to prevent basic security issues
func SanitizeInput(input string) string {
	// Remove potentially dangerous characters
	input = strings.ReplaceAll(input, "<", "")
	input = strings.ReplaceAll(input, ">", "")
	input = strings.ReplaceAll(input, "\"", "")
	input = strings.ReplaceAll(input, "'", "")
	input = strings.ReplaceAll(input, "&", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}
