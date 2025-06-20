from flask import Flask, render_template, jsonify, request, Response
from flask_cors import CORS
import json
import random
from datetime import datetime, date
import hashlib
from difflib import SequenceMatcher
import requests

app = Flask(__name__)
CORS(app)

# Configure static files
app.static_folder = 'static'

# Load player data
with open('players.json', 'r') as f:
    PLAYERS_DATA = json.load(f)

# Create a dictionary for faster lookup by username
PLAYERS_DICT = {player['player_username']: player for player in PLAYERS_DATA}


def get_champion_img(champion):
    curated = champion.replace("\'", "").replace(" ", "").replace(".", "")
    match curated:
        case "KaiSa":
            curated = "Kaisa"
        case "Wukong":
            curated = "MonkeyKing"
        case "RenataGlasc":
            curated = "Renata"
    return f"https://ddragon.leagueoflegends.com/cdn/img/champion/centered/{curated}_0.jpg"


def get_team_logo(team_name):
    """Get team logo URL - now uses URLs from player data instead of local files"""
    # This function is kept for compatibility but will use the URL from player_team_media_url
    # The actual URL is now provided by the scraper via PandaScore API
    return ""  # Will be overridden by the URL from player data


def get_country_flag(country):
    """Get flag emoji for a country"""
    # Map of country names to their flag emojis
    country_flags = {
        "Belgium": "🇧🇪",
        "Canada": "🇨🇦",
        "Czech Republic": "🇨🇿",
        "Denmark": "🇩🇰",
        "France": "🇫🇷",
        "Germany": "🇩🇪",
        "Greece": "🇬🇷",
        "Lithuania": "🇱🇹",
        "Morocco": "🇲🇦",
        "Poland": "🇵🇱",
        "Slovenia": "🇸🇮",
        "South Korea": "🇰🇷",
        "Spain": "🇪🇸",
        "Sweden": "🇸🇪",
        "Turkey": "🇹🇷",
        # Additional common countries for future players
        "United Kingdom": "🇬🇧",
        "China": "🇨🇳",
        "Japan": "🇯🇵",
        "Taiwan": "🇹🇼",
        "Vietnam": "🇻🇳",
        "United States": "🇺🇸",
        "Brazil": "🇧🇷",
        "Australia": "🇦🇺",
        "Netherlands": "🇳🇱",
        "Norway": "🇳🇴",
        "Finland": "🇫🇮",
        "Austria": "🇦🇹",
        "Switzerland": "🇨🇭",
        "Italy": "🇮🇹",
        "Portugal": "🇵🇹",
        "Croatia": "🇭🇷",
        "Serbia": "🇷🇸",
        "Hungary": "🇭🇺",
        "Romania": "🇷🇴",
        "Bulgaria": "🇧🇬"
    }
    return country_flags.get(country, "🏳️")


# Global variable to store custom daily player override
_daily_player_override = None

def get_daily_player():
    """Get the same player for all users on a given day"""
    global _daily_player_override
    
    # If there's an override set, use it
    if _daily_player_override:
        return _daily_player_override
    
    # Use today's date as seed
    today = date.today().isoformat()
    seed = int(hashlib.md5(today.encode()).hexdigest(), 16)
    random.seed(seed)
    return random.choice(PLAYERS_DATA)


def get_player_suggestions(query, limit=3):
    """Get player suggestions based on query"""
    if not query:
        return []

    query_lower = query.lower()
    suggestions = []

    # First, check usernames that start with the query
    starts_with = [p for p in PLAYERS_DATA if p['player_username'].lower().startswith(query_lower)]
    suggestions.extend(starts_with[:limit])

    if len(suggestions) < limit:
        # Then, check usernames that contain the query
        contains = [p for p in PLAYERS_DATA
                    if query_lower in p['player_username'].lower()
                    and p not in suggestions]
        suggestions.extend(contains[:limit - len(suggestions)])

    if len(suggestions) < limit:
        # Finally, find the closest matches using SequenceMatcher
        all_players = [(p, SequenceMatcher(None, query_lower, p['player_username'].lower()).ratio())
                       for p in PLAYERS_DATA if p not in suggestions]
        all_players.sort(key=lambda x: x[1], reverse=True)

        for player, _ in all_players[:limit - len(suggestions)]:
            suggestions.append(player)

    return [{'username': p['player_username'], 'team': p['player_team']} for p in suggestions[:limit]]


def compare_players(guess, target):
    """Compare guessed player with target player"""
    result = {
        'username': guess['player_username'],
        'team': {
            'value': guess['player_team'],
            'logo': guess['player_team_media_url'],
            'status': 'correct' if guess['player_team'] == target['player_team']
            else 'partial' if guess['player_league'] == target['player_league']
            else 'incorrect'
        },
        'age': {
            'value': guess['player_age'],
            'status': 'correct' if guess['player_age'] == target['player_age'] else 'incorrect',
            'direction': 'higher' if target['player_age'] > guess['player_age'] else 'lower'
        },
        'role': {
            'value': guess['player_role'],
            'status': 'correct' if guess['player_role'] == target['player_role'] else 'incorrect'
        },
        'country': {
            'value': guess['player_country'],
            'flag': get_country_flag(guess['player_country']),
            'status': 'correct' if guess['player_country'] == target['player_country']
            else 'partial' if guess['player_country_continent'] == target['player_country_continent']
            else 'incorrect'
        },
        'kda': {
            'value': round(guess['kda_ratio'], 2),
            'status': 'correct' if guess['kda_ratio'] == target['kda_ratio'] else 'incorrect',
            'direction': 'higher' if target['kda_ratio'] > guess['kda_ratio'] else 'lower'
        },
        'champion': {
            'value': guess['player_most_played_champion'],
            'image': get_champion_img(guess['player_most_played_champion']),
            'status': 'correct' if guess['player_most_played_champion'] == target[
                'player_most_played_champion'] else 'incorrect'
        },
        'is_correct': guess['player_username'] == target['player_username']
    }
    return result


@app.route('/')
def index():
    return render_template('index.html')


@app.route('/api/suggestions')
def get_suggestions():
    query = request.args.get('q', '')
    suggestions = get_player_suggestions(query)
    return jsonify(suggestions)


@app.route('/api/guess', methods=['POST'])
def make_guess():
    data = request.get_json()
    username = data.get('username')

    if username not in PLAYERS_DICT:
        return jsonify({'error': 'Player not found'}), 404

    guessed_player = PLAYERS_DICT[username]
    daily_player = get_daily_player()

    result = compare_players(guessed_player, daily_player)
    return jsonify(result)


@app.route('/api/debug/answer')
def debug_answer():
    """Debug endpoint to see today's answer"""
    return jsonify(get_daily_player())


@app.route('/api/proxy/team-image')
def proxy_team_image():
    """Proxy endpoint for team images to avoid CORS issues"""
    image_url = request.args.get('url')
    
    if not image_url:
        return jsonify({'error': 'URL parameter is required'}), 400
    
    try:
        # Fetch the image from the original URL
        response = requests.get(image_url, timeout=10, headers={
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
        })
        
        if response.status_code == 200:
            # Return the image with appropriate headers
            return Response(
                response.content,
                mimetype=response.headers.get('Content-Type', 'image/png'),
                headers={
                    'Cache-Control': 'public, max-age=3600',  # Cache for 1 hour
                    'Access-Control-Allow-Origin': '*'
                }
            )
        else:
            return jsonify({'error': 'Failed to fetch image'}), response.status_code
            
    except requests.RequestException as e:
        return jsonify({'error': f'Request failed: {str(e)}'}), 500


@app.route('/api/reroll', methods=['POST'])
def reroll_daily_player():
    """Reroll the daily player choice"""
    global _daily_player_override
    
    # Select a new random player
    new_player = random.choice(PLAYERS_DATA)
    _daily_player_override = new_player
    
    return jsonify({
        'message': 'Daily player rerolled successfully',
        'new_player': new_player['player_username']
    })


if __name__ == '__main__':
    from waitress import serve

    serve(app, host='0.0.0.0', port=8080)