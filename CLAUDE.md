# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is "Prodle" - a League of Legends Wordle-style game where players guess professional LEC (League of Legends European Championship) players based on their attributes. The game displays clues about team, age, role, country, KDA ratio, and most played champion.

## Architecture

- **Flask Backend** (`main.py`): Serves the web application and API endpoints
  - `/api/suggestions` - Returns player suggestions based on partial username input
  - `/api/guess` - Processes player guesses and returns comparison results
  - `/api/debug/answer` - Debug endpoint showing today's answer
- **Frontend**: Single-page HTML application with vanilla JavaScript (`templates/index.html`)
- **Data Source**: JSON file containing LEC player data (`players.json`)
- **Data Scraper**: Python script to fetch player data from Leaguepedia (`lec_players_scraper_improved.py`)

## Key Components

### Daily Player Selection
Uses deterministic random selection based on today's date (MD5 hash of date as seed) to ensure all players get the same target player on the same day.

### Player Comparison Logic
The `compare_players()` function returns detailed comparison results with status indicators:
- `correct`: Exact match
- `partial`: Partial match (same league for team, same continent for country)
- `incorrect`: No match, with directional hints for age/KDA

### Player Data Structure
Each player entry contains:
- Basic info: username, real name, team, role, country, age
- Statistics: KDA ratio, most played champion, number of clubs
- Media URLs: player and team logos
- Metadata: league, continent

## Running the Application

```bash
python main.py
```

The Flask app runs in debug mode by default and serves on localhost:5000.

## Data Management

To update player data, run the scraper:
```bash
python lec_players_scraper_improved.py
```

This fetches current LEC player data from Leaguepedia and saves it to `players.json`. The scraper includes rate limiting and error handling for the API calls.

## Dependencies

The project uses:
- Flask with CORS support for the web server
- mwrogue.esports_client for Leaguepedia API access
- Standard Python libraries for data processing

## Game Mechanics

- Players have 10 attempts to guess the correct player
- Each guess shows comparison results across 6 attributes
- The game resets daily with a new target player
- Visual feedback uses color coding (green=correct, yellow=partial, gray=incorrect)