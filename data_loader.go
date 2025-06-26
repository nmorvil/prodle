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
	teamImages    map[string]string
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

	// Load team images
	if err := LoadTeamImages(); err != nil {
		return fmt.Errorf("failed to load team images: %v", err)
	}

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

	allPlayers = players
	log.Printf("Loaded %d players from prodle.json", len(allPlayers))
	return nil
}

// LoadTeamImages reads and parses the img_mapping.json file
func LoadTeamImages() error {
	data, err := os.ReadFile("data/img_mapping.json")
	if err != nil {
		return fmt.Errorf("failed to read img_mapping.json: %v", err)
	}

	var images map[string]string
	if err := json.Unmarshal(data, &images); err != nil {
		return fmt.Errorf("failed to parse img_mapping.json: %v", err)
	}

	teamImages = images
	log.Printf("Loaded %d team image mappings from img_mapping.json", len(teamImages))
	return nil
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

// GetAllPlayers returns all loaded players
func GetAllPlayers() []Player {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil
	}

	// Return a copy to prevent external modification
	result := make([]Player, len(allPlayers))
	copy(result, allPlayers)
	return result
}

// GetAllPlayerNames returns all player usernames for autocomplete
func GetAllPlayerNames() []string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil
	}

	// Return a copy to prevent external modification
	result := make([]string, len(playerNames))
	copy(result, playerNames)
	return result
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

// GetTeamImage returns the image filename for a team
func GetTeamImage(teamName string) (string, bool) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return "", false
	}

	image, exists := teamImages[teamName]
	return image, exists
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

// FilterPlayersByName returns players whose names match the search query (for autocomplete)
func FilterPlayersByName(query string, limit int) []string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded || query == "" {
		return nil
	}

	query = strings.ToLower(query)
	var matches []string

	for _, name := range playerNames {
		if strings.Contains(strings.ToLower(name), query) {
			matches = append(matches, name)
			if len(matches) >= limit {
				break
			}
		}
	}

	return matches
}

// GetDataStats returns statistics about the loaded data
func GetDataStats() map[string]interface{} {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return map[string]interface{}{
			"loaded": false,
		}
	}

	return map[string]interface{}{
		"loaded":        true,
		"total_players": len(allPlayers),
		"total_teams":   len(GetAllTeams()),
		"total_leagues": len(GetAllLeagues()),
		"total_roles":   len(GetAllRoles()),
		"team_images":   len(teamImages),
	}
}
