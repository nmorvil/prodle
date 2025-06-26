# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a League of Legends "Prodle" game - a Wordle-style guessing game for League of Legends esports players. Players guess professional players based on various attributes like team, league, role, country, age, and statistics. The game features a sophisticated scoring system, session management, and leaderboard functionality.

## Development Commands

```bash
# Initialize dependencies
go mod tidy

# Run the application
go run main.go

# Build the application
go build

# Format code
go fmt ./...

# Check for issues
go vet

# Run tests (when they exist)
go test ./...
```

## Architecture

### File Structure and Responsibilities

**Core Files:**
- `main.go` - Entry point (currently minimal, needs web server implementation)
- `models.go` - Data structures for Player, GameSession, GuessResult, LeaderboardEntry
- `database.go` - SQLite database layer with leaderboard functionality
- `data_loader.go` - JSON data loading and in-memory caching
- `game_logic.go` - Session management, game state, and validation logic
- `utils.go` - Utility functions including scoring and champion image URLs

**Data Files:**
- `data/prodle.json` - 3000+ LoL esports players with comprehensive stats
- `data/img_mapping.json` - Team name to logo image mapping
- `assets/` - Static assets (team logos, fonts, backgrounds)

**Dependencies:**
- `github.com/mattn/go-sqlite3` - SQLite database driver

### Core Game Architecture

**Session Management:**
- 20 random players per session
- 2-minute time limit per player
- Maximum 6 guesses per player
- Thread-safe in-memory session storage with automatic cleanup

**Scoring System:**
- Time-based scoring: 5000 points decreasing linearly to 1000 over 2 minutes
- Formula: `points = 5000 - (4000 * elapsedSeconds / 120)`
- Wrong guess penalty: -100 points per incorrect guess
- Completion bonus: +1000 points for finishing all 20 players

**Comparison Logic:**
- **Exact Match** (Green): All attributes match perfectly
- **Partial Match** (Yellow): 
  - Team: Same league, different team
  - Country: Same continent, different country
- **Higher/Lower**: Age, KDA ratio, number of clubs, statistical averages
- **Wrong**: No match

## Key Functions by Layer

### Database Layer (`database.go`)
```go
InitDatabase() - Initialize SQLite with leaderboard table
AddToLeaderboard(username, score) - Add score entry
AddToLeaderboardFromSession(username, session) - Add with full session data
GetTop10Scores() - Retrieve top 10 leaderboard entries
GetFormattedTop10() - Get formatted leaderboard for display
```

### Data Layer (`data_loader.go`)
```go
InitializeGameData() - Load all JSON data into memory
GetRandomPlayers(count) - Select random players for session
GetPlayerByName(name) - Case-insensitive player lookup
GetAllPlayerNames() - For autocomplete functionality
FilterPlayersByName(query, limit) - Search with query filtering
```

### Game Logic (`game_logic.go`)
```go
CreateNewSession() - Create new game with 20 random players
ValidateGuess(session, playerName) - Process guess with detailed comparisons
CheckCorrectGuess(playerName) - Simple boolean check for correctness
IsGameOver() - Check time limits and completion status
MoveToNextPlayer() - Advance to next player with state reset
```

### Utilities (`utils.go`)
```go
GetChampionImg(champion) - Generate League of Legends champion image URLs
CalculatePlayerScore(elapsedSeconds, wrongGuesses) - Core scoring logic
ValidatePlayerGuess(guess) - Input validation
SanitizeInput(input) - Security sanitization
```

## Game Flow

1. **Session Creation**: `CreateNewSession()` selects 20 random players
2. **Player Guessing**: Players make guesses validated by `ValidateGuess()`
3. **Attribute Comparison**: Detailed comparison with color-coded feedback
4. **Score Calculation**: Time and accuracy-based scoring per player
5. **Progression**: Move to next player or complete session
6. **Leaderboard**: Final scores saved to SQLite database

## Data Models

**Player Structure:**
- Basic info: username, real name, team, league, country, age, role
- Statistics: KDA ratio, avg kills/deaths/assists, games played
- Metadata: number of clubs, most played champion

**Session Management:**
- Session ID for tracking
- Current player index and start times
- Guess history with detailed comparisons
- Cumulative scoring and completion status

## Champion Image Integration

The `GetChampionImg()` function generates champion image URLs from Riot's Data Dragon CDN:
- Handles special cases: Kai'Sa → Kaisa, Wukong → MonkeyKing, Renata Glasc → Renata
- URL format: `https://ddragon.leagueoflegends.com/cdn/img/champion/centered/{champion}_0.jpg`

## Development Guidelines

1. **Layer Separation**: Keep SQL queries only in `database.go`
2. **Thread Safety**: All session operations use proper mutex locking
3. **Error Handling**: Comprehensive error handling with descriptive messages
4. **Input Validation**: All user inputs are sanitized and validated
5. **Scoring Integrity**: Score calculations are deterministic and traceable

## Next Implementation Steps

1. **Web Server**: Implement HTTP handlers for game endpoints
2. **Frontend**: Create HTML templates and JavaScript for game interface
3. **API Layer**: REST endpoints for session management and guessing
4. **Real-time Updates**: WebSocket support for live score updates
5. **Authentication**: User accounts and persistent statistics