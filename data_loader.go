package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// In-memory data storage
var (
	allPlayers    []Player
	playerNames   []string
	playersByName map[string]Player
	dataLoaded    bool
	dataMutex     sync.RWMutex
)

// InitializeGameData loads all game data into memory
func InitializeGameData() error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	if dataLoaded {
		return nil // Already loaded
	}

	// Load players
	if err := LoadPlayers(); err != nil {
		return fmt.Errorf("failed to load players: %v", err)
	}

	// Team images are now loaded directly from team names
	// No need to load separate mapping file

	// Initialize player lookup structures
	initializePlayerLookup()

	dataLoaded = true
	log.Printf("Game data initialized successfully: %d players loaded", len(allPlayers))
	return nil
}

// LoadPlayers reads and parses the prodle.json file
func LoadPlayers() error {
	data, err := os.ReadFile("data/prodle.json")
	if err != nil {
		return fmt.Errorf("failed to read prodle.json: %v", err)
	}

	var players []Player
	if err := json.Unmarshal(data, &players); err != nil {
		return fmt.Errorf("failed to parse prodle.json: %v", err)
	}

	// Populate compatibility fields for backward compatibility
	for i := range players {
		populateCompatibilityFields(&players[i])
	}

	allPlayers = players
	log.Printf("Loaded %d players from prodle.json", len(allPlayers))
	return nil
}

// populateCompatibilityFields fills in the legacy field names for backward compatibility
func populateCompatibilityFields(player *Player) {
	// Map new fields to legacy field names
	player.PlayerUsername = player.ID
	player.PlayerName = player.ID
	player.PlayerTeam = player.Team
	player.PlayerLeague = player.League
	player.NumberOfClubs = len(player.TeamsPlayed)
	player.PlayerCountry = player.Nationality
	player.PlayerCountryContinent = player.Continent
	player.PlayerRole = player.Role

	// Calculate age from year of birth
	currentYear := time.Now().Year()
	if player.YearOfBirth > 0 {
		player.PlayerAge = currentYear - player.YearOfBirth
	}

	// Legacy fields remain empty/zero as they're not in the new JSON structure
	player.PlayerMediaURL = ""
	player.PlayerTeamMediaURL = ""
	player.PlayerMostPlayedChampion = ""
	player.AvgKills = 0.0
	player.AvgDeaths = 0.0
	player.AvgAssists = 0.0
	player.KDARatio = 0.0
	player.GamesPlayed = 0
}

// initializePlayerLookup creates lookup structures for quick player access
func initializePlayerLookup() {
	playersByName = make(map[string]Player)
	playerNames = make([]string, 0, len(allPlayers))

	for _, player := range allPlayers {
		// Use username as the primary key for lookups
		key := strings.ToLower(player.PlayerUsername)
		playersByName[key] = player
		playerNames = append(playerNames, player.PlayerUsername)

		// Also add real name as alternative lookup if different from username
		if player.PlayerName != "" && strings.ToLower(player.PlayerName) != key {
			realNameKey := strings.ToLower(player.PlayerName)
			playersByName[realNameKey] = player
		}
	}

	// Sort player names for autocomplete
	sort.Strings(playerNames)
}

// GetPlayerByName returns a player by their username or real name (case-insensitive)
func GetPlayerByName(name string) (*Player, bool) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil, false
	}

	player, exists := playersByName[strings.ToLower(name)]
	return &player, exists
}

// GetAllTeams returns all unique team names
func GetAllTeams() []string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil
	}

	teamSet := make(map[string]bool)
	for _, player := range allPlayers {
		teamSet[player.PlayerTeam] = true
	}

	teams := make([]string, 0, len(teamSet))
	for team := range teamSet {
		teams = append(teams, team)
	}

	sort.Strings(teams)
	return teams
}

// GetAllLeagues returns all unique league names
func GetAllLeagues() []string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil
	}

	leagueSet := make(map[string]bool)
	for _, player := range allPlayers {
		leagueSet[player.PlayerLeague] = true
	}

	leagues := make([]string, 0, len(leagueSet))
	for league := range leagueSet {
		leagues = append(leagues, league)
	}

	sort.Strings(leagues)
	return leagues
}

// GetAllRoles returns all unique player roles
func GetAllRoles() []string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil
	}

	roleSet := make(map[string]bool)
	for _, player := range allPlayers {
		roleSet[player.PlayerRole] = true
	}

	roles := make([]string, 0, len(roleSet))
	for role := range roleSet {
		roles = append(roles, role)
	}

	sort.Strings(roles)
	return roles
}

// GetRandomPlayers returns a random selection of players for a game session
func GetRandomPlayers(count int) ([]Player, error) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil, fmt.Errorf("game data not loaded")
	}

	if count > len(allPlayers) {
		count = len(allPlayers)
	}

	// Create a copy of all players and shuffle
	players := make([]Player, len(allPlayers))
	copy(players, allPlayers)

	// Shuffle using Fisher-Yates algorithm
	rand.Seed(time.Now().UnixNano())
	for i := len(players) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		players[i], players[j] = players[j], players[i]
	}

	return players[:count], nil
}

// Difficulty level constants
const (
	DifficultyFacile    = "facile"
	DifficultyMoyen     = "moyen"
	DifficultyDifficile = "difficile"
)

// parseRankingToIntForFilter converts ranking string to integer for filtering
func parseRankingToIntForFilter(ranking string) int {
	if ranking == "" {
		return 999 // Default high value for invalid rankings
	}

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

// GetPlayersByDifficulty returns players filtered by difficulty level
func GetPlayersByDifficulty(difficulty string) []Player {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil
	}

	switch difficulty {
	case DifficultyFacile:
		return getFacilePlayersUnsafe()
	case DifficultyMoyen:
		return getMoyenPlayersUnsafe()
	case DifficultyDifficile:
		return getDifficilePlayersUnsafe()
	default:
		// Return a copy to prevent external modification
		result := make([]Player, len(allPlayers))
		copy(result, allPlayers)
		return result
	}
}

// IsPlayerInDifficulty checks if a specific player is available in the given difficulty
func IsPlayerInDifficulty(player *Player, difficulty string) bool {
	if player == nil {
		return false
	}

	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return false
	}

	switch difficulty {
	case DifficultyFacile:
		return isPlayerInFacile(*player)
	case DifficultyMoyen:
		return isPlayerInMoyen(*player)
	case DifficultyDifficile:
		return isPlayerInDifficile(*player)
	default:
		return true // If no specific difficulty, all players are valid
	}
}

// Helper functions for individual player difficulty checks
func isPlayerInFacile(player Player) bool {
	league := player.League
	rank := parseRankingToIntForFilter(player.LastSplitResult)

	switch league {
	case "LoL EMEA Championship": // LEC
		return true
	case "La Ligue Française": // LFL
		return rank <= 5
	case "LoL Champions Korea": // LCK
		return rank <= 5
	default:
		return false
	}
}

func isPlayerInMoyen(player Player) bool {
	league := player.League
	rank := parseRankingToIntForFilter(player.LastSplitResult)

	switch league {
	case "LoL EMEA Championship", "La Ligue Française", "LoL Champions Korea": // LEC, LFL, LCK
		return true
	case "Tencent LoL Pro League": // LPL
		return rank <= 6
	default:
		return false
	}
}

func isPlayerInDifficile(player Player) bool {
	league := player.League
	rank := parseRankingToIntForFilter(player.LastSplitResult)

	switch league {
	case "League of Legends Championship of The Americas North": // LTAN
		return rank <= 4
	case "LoL Champions Korea", "LoL EMEA Championship", "La Ligue Française": // LCK, LEC, LFL
		return true
	case "Tencent LoL Pro League": // LPL
		return rank <= 10
	case "League of Legends Championship Pacific": // LCP
		return rank <= 3
	default:
		return false
	}
}

// getFacilePlayers returns players for Facile difficulty
// LEC (all), LFL (top 5), LCK (top 5)
func getFacilePlayers() []Player {
	dataMutex.RLock()
	defer dataMutex.RUnlock()
	return getFacilePlayersUnsafe()
}

// getFacilePlayersUnsafe returns players for Facile difficulty (assumes lock is already held)
func getFacilePlayersUnsafe() []Player {
	var result []Player

	for _, player := range allPlayers {
		league := player.League
		rank := parseRankingToIntForFilter(player.LastSplitResult)

		switch league {
		case "LoL EMEA Championship": // LEC
			result = append(result, player)
		case "La Ligue Française": // LFL
			if rank <= 5 {
				result = append(result, player)
			}
		case "LoL Champions Korea": // LCK
			if rank <= 5 {
				result = append(result, player)
			}
		}
	}

	return result
}

// getMoyenPlayers returns players for Moyen difficulty
// LEC (all), LFL (all), LCK (all), LPL (top 6)
func getMoyenPlayers() []Player {
	dataMutex.RLock()
	defer dataMutex.RUnlock()
	return getMoyenPlayersUnsafe()
}

// getMoyenPlayersUnsafe returns players for Moyen difficulty (assumes lock is already held)
func getMoyenPlayersUnsafe() []Player {
	var result []Player

	for _, player := range allPlayers {
		league := player.League
		rank := parseRankingToIntForFilter(player.LastSplitResult)

		switch league {
		case "LoL EMEA Championship": // LEC
			result = append(result, player)
		case "La Ligue Française": // LFL
			result = append(result, player)
		case "LoL Champions Korea": // LCK
			result = append(result, player)
		case "Tencent LoL Pro League": // LPL
			if rank <= 6 {
				result = append(result, player)
			}
		}
	}

	return result
}

// getDifficilePlayers returns players for Difficile difficulty
// LTAN (top 4), LCK (all), LPL (top 10), LEC (all), LFL (all), LCP (top 3)
func getDifficilePlayers() []Player {
	dataMutex.RLock()
	defer dataMutex.RUnlock()
	return getDifficilePlayersUnsafe()
}

// getDifficilePlayersUnsafe returns players for Difficile difficulty (assumes lock is already held)
func getDifficilePlayersUnsafe() []Player {
	var result []Player

	for _, player := range allPlayers {
		league := player.League
		rank := parseRankingToIntForFilter(player.LastSplitResult)

		switch league {
		case "League of Legends Championship of The Americas North": // LTAN
			if rank <= 4 {
				result = append(result, player)
			}
		case "LoL Champions Korea": // LCK
			result = append(result, player)
		case "Tencent LoL Pro League": // LPL
			if rank <= 10 {
				result = append(result, player)
			}
		case "LoL EMEA Championship": // LEC
			result = append(result, player)
		case "La Ligue Française": // LFL
			result = append(result, player)
		case "League of Legends Championship Pacific": // LCP
			if rank <= 3 {
				result = append(result, player)
			}
		}
	}

	return result
}

// GetRandomPlayersByDifficulty returns random players filtered by difficulty
func GetRandomPlayersByDifficulty(count int, difficulty string) ([]Player, error) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil, fmt.Errorf("game data not loaded")
	}

	filteredPlayers := GetPlayersByDifficulty(difficulty)
	if len(filteredPlayers) == 0 {
		return nil, fmt.Errorf("no players found for difficulty %s", difficulty)
	}

	if count > len(filteredPlayers) {
		count = len(filteredPlayers)
	}

	// Create a copy and shuffle
	players := make([]Player, len(filteredPlayers))
	copy(players, filteredPlayers)

	// Shuffle using Fisher-Yates algorithm
	rand.Seed(time.Now().UnixNano())
	for i := len(players) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		players[i], players[j] = players[j], players[i]
	}

	return players[:count], nil
}

// FilterPlayersByNameAndDifficulty returns player names filtered by search query and difficulty
func FilterPlayersByNameAndDifficulty(query string, difficulty string, limit int) []string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil
	}

	filteredPlayers := GetPlayersByDifficulty(difficulty)

	// If query is empty, return all players from the difficulty
	if query == "" {
		var matches []string
		for _, player := range filteredPlayers {
			matches = append(matches, player.ID)
			if len(matches) >= limit {
				break
			}
		}
		return matches
	}

	query = strings.ToLower(query)

	var matches []string
	for _, player := range filteredPlayers {
		if strings.Contains(strings.ToLower(player.ID), query) {
			matches = append(matches, player.ID)
			if len(matches) >= limit {
				break
			}
		}
	}

	return matches
}

// GetDifficultyInfo returns information about each difficulty level
func GetDifficultyInfo() map[string]map[string]interface{} {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil
	}

	info := make(map[string]map[string]interface{})

	// Facile info
	facilePlayers := getFacilePlayers()
	info[DifficultyFacile] = map[string]interface{}{
		"name":        "Facile",
		"playerCount": len(facilePlayers),
		"leagues":     "LEC, Top 5 LFL, Top 5 LCK",
		"description": "",
	}

	// Moyen info
	moyenPlayers := getMoyenPlayers()
	info[DifficultyMoyen] = map[string]interface{}{
		"name":        "Moyen",
		"playerCount": len(moyenPlayers),
		"leagues":     "LEC, LFL, LCK, Top 6 LPL",
		"description": "",
	}

	// Difficile info
	difficilePlayers := getDifficilePlayers()
	info[DifficultyDifficile] = map[string]interface{}{
		"name":        "Difficile",
		"playerCount": len(difficilePlayers),
		"leagues":     "Top 4 LTAN, LCK, Top 10 LPL, LEC, LFL, Top 3 LCP",
		"description": "",
	}

	return info
}
