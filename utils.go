package main

import (
	"fmt"
	"strings"
)

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

func CalculatePlayerPoints(totalElapsedSeconds, wrongGuesses int) int {

	basePoints := 5000

	timeProgress := float64(totalElapsedSeconds) / float64(TotalGameTime)
	if timeProgress > 1.0 {
		timeProgress = 1.0
	}

	points := int(float64(basePoints) * (1.0 - 0.7*timeProgress))

	penalty := wrongGuesses * 100
	points -= penalty

	if points < 300 {
		points = 300
	}

	return points
}

func CalculateGameScore(totalElapsedSeconds, totalWrongGuesses, playersFound int) int {

	baseScore := playersFound * 3000
	timePenalty := totalElapsedSeconds * 10
	guessPenalty := totalWrongGuesses * 50

	score := baseScore - timePenalty - guessPenalty
	if score < 0 {
		score = 0
	}

	return score
}

func CalculatePoints(elapsedSeconds int) int {

	return CalculatePlayerPoints(elapsedSeconds, 0)
}

func CalculatePlayerScore(elapsedSeconds int, wrongGuesses int) int {
	basePoints := CalculatePoints(elapsedSeconds)

	penalty := wrongGuesses * 100

	finalScore := basePoints - penalty

	if finalScore < 0 {
		finalScore = 0
	}

	return finalScore
}

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

func SanitizeInput(input string) string {

	input = strings.ReplaceAll(input, "<", "")
	input = strings.ReplaceAll(input, ">", "")
	input = strings.ReplaceAll(input, "\"", "")
	input = strings.ReplaceAll(input, "'", "")
	input = strings.ReplaceAll(input, "&", "")

	input = strings.TrimSpace(input)

	return input
}
