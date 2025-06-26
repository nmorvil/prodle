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

// Contains checks if a string slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsIgnoreCase checks if a string slice contains a specific string (case-insensitive)
func ContainsIgnoreCase(slice []string, item string) bool {
	item = strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// TrimAndLower trims whitespace and converts to lowercase
func TrimAndLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

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

// CalculatePoints calculates points for a single player based on elapsed time
// Formula: points = 5000 - (4000 * elapsedSeconds / 120)
// Linear decrease from 5000 points at 0 seconds to 1000 points at 2 minutes (120 seconds)
func CalculatePoints(elapsedSeconds int) int {
	// Ensure minimum of 1000 points after 2 minutes
	if elapsedSeconds >= 120 {
		return 1000
	}

	// Linear decrease: 5000 - (4000 * elapsedSeconds / 120)
	points := 5000 - (4000 * elapsedSeconds / 120)

	// Ensure points don't go below 1000
	if points < 1000 {
		points = 1000
	}

	return points
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

// CalculateScore calculates the game score based on guesses and time (legacy function)
// Deprecated: Use CalculatePlayerScore instead
func CalculateScore(guessCount int, durationSeconds int, maxGuesses int) int {
	if guessCount > maxGuesses {
		return 0 // No score if exceeded max guesses
	}

	wrongGuesses := guessCount - 1 // First guess doesn't count as wrong
	if wrongGuesses < 0 {
		wrongGuesses = 0
	}

	return CalculatePlayerScore(durationSeconds, wrongGuesses)
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
