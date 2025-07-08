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

var (
	allPlayers    []Player
	playerNames   []string
	playersByName map[string]Player
	dataLoaded    bool
	dataMutex     sync.RWMutex
)

func InitializeGameData() error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	if dataLoaded {
		return nil
	}

	if err := LoadPlayers(); err != nil {
		return fmt.Errorf("failed to load players: %v", err)
	}

	initializePlayerLookup()

	dataLoaded = true
	log.Printf("Game data initialized successfully: %d players loaded", len(allPlayers))
	return nil
}

func LoadPlayers() error {
	data, err := os.ReadFile("data/prodle.json")
	if err != nil {
		return fmt.Errorf("failed to read prodle.json: %v", err)
	}

	var players []Player
	if err := json.Unmarshal(data, &players); err != nil {
		return fmt.Errorf("failed to parse prodle.json: %v", err)
	}

	for i := range players {
		populateCompatibilityFields(&players[i])
	}

	allPlayers = players
	log.Printf("Loaded %d players from prodle.json", len(allPlayers))
	return nil
}

func populateCompatibilityFields(player *Player) {

	player.PlayerUsername = player.ID
	player.PlayerName = player.ID
	player.PlayerTeam = player.Team
	player.PlayerLeague = player.League
	player.NumberOfClubs = len(player.TeamsPlayed)
	player.PlayerCountry = player.Nationality
	player.PlayerCountryContinent = player.Continent
	player.PlayerRole = player.Role

	currentYear := time.Now().Year()
	if player.YearOfBirth > 0 {
		player.PlayerAge = currentYear - player.YearOfBirth
	}

	player.PlayerMediaURL = ""
	player.PlayerTeamMediaURL = ""
	player.PlayerMostPlayedChampion = ""
	player.AvgKills = 0.0
	player.AvgDeaths = 0.0
	player.AvgAssists = 0.0
	player.KDARatio = 0.0
	player.GamesPlayed = 0
}

func initializePlayerLookup() {
	playersByName = make(map[string]Player)
	playerNames = make([]string, 0, len(allPlayers))

	for _, player := range allPlayers {

		key := strings.ToLower(player.PlayerUsername)
		playersByName[key] = player
		playerNames = append(playerNames, player.PlayerUsername)

		if player.PlayerName != "" && strings.ToLower(player.PlayerName) != key {
			realNameKey := strings.ToLower(player.PlayerName)
			playersByName[realNameKey] = player
		}
	}

	sort.Strings(playerNames)
}

func GetPlayerByName(name string) (*Player, bool) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil, false
	}

	player, exists := playersByName[strings.ToLower(name)]
	return &player, exists
}

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

func GetRandomPlayers(count int) ([]Player, error) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil, fmt.Errorf("game data not loaded")
	}

	if count > len(allPlayers) {
		count = len(allPlayers)
	}

	players := make([]Player, len(allPlayers))
	copy(players, allPlayers)

	rand.Seed(time.Now().UnixNano())
	for i := len(players) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		players[i], players[j] = players[j], players[i]
	}

	return players[:count], nil
}

const (
	DifficultyFacile    = "facile"
	DifficultyMoyen     = "moyen"
	DifficultyDifficile = "difficile"
)

func parseRankingToIntForFilter(ranking string) int {
	if ranking == "" {
		return 999
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

		result := make([]Player, len(allPlayers))
		copy(result, allPlayers)
		return result
	}
}

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
		return true
	}
}

func isPlayerInFacile(player Player) bool {
	league := player.League
	rank := parseRankingToIntForFilter(player.LastSplitResult)

	switch league {
	case "LoL EMEA Championship":
		return true
	case "La Ligue Française":
		return rank <= 5
	case "LoL Champions Korea":
		return rank <= 5
	default:
		return false
	}
}

func isPlayerInMoyen(player Player) bool {
	league := player.League
	rank := parseRankingToIntForFilter(player.LastSplitResult)

	switch league {
	case "LoL EMEA Championship", "La Ligue Française", "LoL Champions Korea":
		return true
	case "Tencent LoL Pro League":
		return rank <= 6
	default:
		return false
	}
}

func isPlayerInDifficile(player Player) bool {
	league := player.League
	rank := parseRankingToIntForFilter(player.LastSplitResult)

	switch league {
	case "League of Legends Championship of The Americas North":
		return rank <= 4
	case "LoL Champions Korea", "LoL EMEA Championship", "La Ligue Française":
		return true
	case "Tencent LoL Pro League":
		return rank <= 10
	case "League of Legends Championship Pacific":
		return rank <= 3
	default:
		return false
	}
}

func getFacilePlayers() []Player {
	dataMutex.RLock()
	defer dataMutex.RUnlock()
	return getFacilePlayersUnsafe()
}

func getFacilePlayersUnsafe() []Player {
	var result []Player

	for _, player := range allPlayers {
		league := player.League
		rank := parseRankingToIntForFilter(player.LastSplitResult)

		switch league {
		case "LoL EMEA Championship":
			result = append(result, player)
		case "La Ligue Française":
			if rank <= 5 {
				result = append(result, player)
			}
		case "LoL Champions Korea":
			if rank <= 5 {
				result = append(result, player)
			}
		}
	}

	return result
}

func getMoyenPlayers() []Player {
	dataMutex.RLock()
	defer dataMutex.RUnlock()
	return getMoyenPlayersUnsafe()
}

func getMoyenPlayersUnsafe() []Player {
	var result []Player

	for _, player := range allPlayers {
		league := player.League
		rank := parseRankingToIntForFilter(player.LastSplitResult)

		switch league {
		case "LoL EMEA Championship":
			result = append(result, player)
		case "La Ligue Française":
			result = append(result, player)
		case "LoL Champions Korea":
			result = append(result, player)
		case "Tencent LoL Pro League":
			if rank <= 6 {
				result = append(result, player)
			}
		}
	}

	return result
}

func getDifficilePlayers() []Player {
	dataMutex.RLock()
	defer dataMutex.RUnlock()
	return getDifficilePlayersUnsafe()
}

func getDifficilePlayersUnsafe() []Player {
	var result []Player

	for _, player := range allPlayers {
		league := player.League
		rank := parseRankingToIntForFilter(player.LastSplitResult)

		switch league {
		case "League of Legends Championship of The Americas North":
			if rank <= 4 {
				result = append(result, player)
			}
		case "LoL Champions Korea":
			result = append(result, player)
		case "Tencent LoL Pro League":
			if rank <= 10 {
				result = append(result, player)
			}
		case "LoL EMEA Championship":
			result = append(result, player)
		case "La Ligue Française":
			result = append(result, player)
		case "League of Legends Championship Pacific":
			if rank <= 3 {
				result = append(result, player)
			}
		}
	}

	return result
}

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

	players := make([]Player, len(filteredPlayers))
	copy(players, filteredPlayers)

	rand.Seed(time.Now().UnixNano())
	for i := len(players) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		players[i], players[j] = players[j], players[i]
	}

	return players[:count], nil
}

func FilterPlayersByNameAndDifficulty(query string, difficulty string, limit int) []string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil
	}

	filteredPlayers := GetPlayersByDifficulty(difficulty)

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

func GetDifficultyInfo() map[string]map[string]interface{} {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if !dataLoaded {
		return nil
	}

	info := make(map[string]map[string]interface{})

	facilePlayers := getFacilePlayers()
	info[DifficultyFacile] = map[string]interface{}{
		"name":        "Facile",
		"playerCount": len(facilePlayers),
		"leagues":     "LEC, Top 5 LFL, Top 5 LCK",
		"description": "",
	}

	moyenPlayers := getMoyenPlayers()
	info[DifficultyMoyen] = map[string]interface{}{
		"name":        "Moyen",
		"playerCount": len(moyenPlayers),
		"leagues":     "LEC, LFL, LCK, Top 6 LPL",
		"description": "",
	}

	difficilePlayers := getDifficilePlayers()
	info[DifficultyDifficile] = map[string]interface{}{
		"name":        "Difficile",
		"playerCount": len(difficilePlayers),
		"leagues":     "Top 4 LTAN, LCK, Top 10 LPL, LEC, LFL, Top 3 LCP",
		"description": "",
	}

	return info
}
