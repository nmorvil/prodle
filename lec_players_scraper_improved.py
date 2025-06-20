#!/usr/bin/env python3

import json
import time
from datetime import datetime
from mwrogue.esports_client import EsportsClient

def calculate_age(birthdate_str):
    """Calculate age from birthdate string"""
    if not birthdate_str:
        return None
    
    try:
        birth_date = datetime.strptime(birthdate_str, "%Y-%m-%d")
        today = datetime.now()
        age = today.year - birth_date.year
        if today.month < birth_date.month or (today.month == birth_date.month and today.day < birth_date.day):
            age -= 1
        return age
    except (ValueError, AttributeError):
        return None

def get_continent_from_country(country):
    """Map country to continent"""
    continent_mapping = {
        # Europe
        "Germany": "Europe", "France": "Europe", "Spain": "Europe", "Italy": "Europe",
        "United Kingdom": "Europe", "Netherlands": "Europe", "Belgium": "Europe",
        "Sweden": "Europe", "Denmark": "Europe", "Norway": "Europe", "Finland": "Europe",
        "Poland": "Europe", "Czech Republic": "Europe", "Austria": "Europe",
        "Switzerland": "Europe", "Greece": "Europe", "Portugal": "Europe",
        "Hungary": "Europe", "Slovenia": "Europe", "Croatia": "Europe",
        "Slovakia": "Europe", "Estonia": "Europe", "Latvia": "Europe",
        "Lithuania": "Europe", "Romania": "Europe", "Bulgaria": "Europe",
        "Serbia": "Europe", "Bosnia and Herzegovina": "Europe", "Albania": "Europe",
        "North Macedonia": "Europe", "Montenegro": "Europe", "Moldova": "Europe",
        "Ukraine": "Europe", "Russia": "Europe", "Belarus": "Europe",
        "Turkey": "Europe", "Iceland": "Europe", "Ireland": "Europe",
        "Luxembourg": "Europe", "Malta": "Europe", "Cyprus": "Europe",
        # Other continents
        "United States": "North America", "Canada": "North America", "Mexico": "North America",
        "South Korea": "Asia", "China": "Asia", "Japan": "Asia", "Taiwan": "Asia",
        "Vietnam": "Asia", "Thailand": "Asia", "Singapore": "Asia", "Malaysia": "Asia",
        "Philippines": "Asia", "Indonesia": "Asia", "India": "Asia", "Pakistan": "Asia",
        "Australia": "Oceania", "New Zealand": "Oceania",
        "Brazil": "South America", "Argentina": "South America", "Chile": "South America",
        "Peru": "South America", "Colombia": "South America", "Venezuela": "South America",
    }
    return continent_mapping.get(country, "Unknown")


def get_kda_stats(site, player_id, year=2025):
    """Get KDA statistics for a player"""
    try:
        time.sleep(1)
        kda_response = site.cargo_client.query(
            tables="ScoreboardPlayers",
            fields="Kills, Deaths, Assists",
            where=f"Link = '{player_id}' AND DateTime_UTC LIKE '{year}%'",
            limit=50
        )
        
        if not kda_response:
            # Try without year filter if no data found
            time.sleep(1)
            kda_response = site.cargo_client.query(
                tables="ScoreboardPlayers",
                fields="Kills, Deaths, Assists",
                where=f"Link = '{player_id}'",
                limit=20
            )
        
        if kda_response:
            # Calculate averages
            total_kills = sum(float(g.get('Kills', 0) or 0) for g in kda_response)
            total_deaths = sum(float(g.get('Deaths', 0) or 0) for g in kda_response)
            total_assists = sum(float(g.get('Assists', 0) or 0) for g in kda_response)
            games = len(kda_response)
            
            if games > 0:
                avg_kills = total_kills / games
                avg_deaths = total_deaths / games if total_deaths > 0 else 1
                avg_assists = total_assists / games
                kda_ratio = (avg_kills + avg_assists) / avg_deaths
                
                return {
                    "avg_kills": round(avg_kills, 1),
                    "avg_deaths": round(avg_deaths, 1),
                    "avg_assists": round(avg_assists, 1),
                    "kda_ratio": round(kda_ratio, 2),
                    "games_played": games
                }
        
        return {
            "avg_kills": 0,
            "avg_deaths": 0,
            "avg_assists": 0,
            "kda_ratio": 0,
            "games_played": 0
        }
        
    except Exception as e:
        print(f"    Error getting KDA: {e}")
        return {
            "avg_kills": 0,
            "avg_deaths": 0,
            "avg_assists": 0,
            "kda_ratio": 0,
            "games_played": 0
        }

def fetch_lec_players_improved():
    """Fetch improved LEC player data with all requested features"""
    site = EsportsClient("lol")
    
    try:
        print("Step 1: Getting LEC 2025 Spring tournament...")
        time.sleep(1)
        tournaments = site.cargo_client.query(
            tables="Tournaments",
            fields="OverviewPage, Name, Year",
            where="Name='LFL 2025 Spring' OR Name ='LCK 2025 Rounds 1-2' OR Name='LEC 2025 Spring'",
            order_by="Year DESC",
            limit=5
        )
        
        print(f"Found {len(tournaments)} LEC 2025 Spring tournaments")
        
        all_players = {}  # Use dict to avoid duplicates
        
        # Get players from each tournament
        for tournament in tournaments:
            print(f"Getting players from {tournament.get('Name', 'Unknown')}...")
            
            try:
                time.sleep(1)
                tournament_players = site.cargo_client.query(
                    tables="TournamentPlayers",
                    fields="Player, Team, Role",
                    where=f"OverviewPage = '{tournament.get('OverviewPage', '')}'",
                    limit=100
                )
                
                # Filter out coaches, managers, analysts - only keep actual players
                player_roles = ['Top', 'Jungle', 'Mid', 'Bot', 'Support', 'ADC', 'Carry']
                filtered_players = [tp for tp in tournament_players if tp.get('Role', '') in player_roles]
                
                print(f"  Found {len(tournament_players)} total, {len(filtered_players)} players (filtered out coaches)")
                
                for tp in filtered_players:
                    player_id = tp.get('Player', '')
                    if player_id and player_id not in all_players:
                        all_players[player_id] = {
                            'tournament_data': tp,
                            'tournament_name': tournament.get('Name', '')
                        }
                
                # Fixed delay between tournaments
                time.sleep(1)
                
            except Exception as e:
                print(f"  Error getting players from {tournament.get('Name', 'Unknown')}: {e}")
                continue
        
        print(f"\nStep 2: Processing {len(all_players)} unique players...")
        
        processed_players = []
        
        for i, (player_id, player_info) in enumerate(all_players.items()):
            print(f"Processing {i+1}/{len(all_players)}: {player_id}")
            
            try:
                # Get full player information
                time.sleep(1)
                player_details = site.cargo_client.query(
                    tables="PlayerRedirects=PR, Players=P",
                    fields="P.Player, P.Name, P.Country, P.Role, P.Birthdate, P.Image",
                    where=f"PR.AllName = '{player_id}'",
                    join_on="PR.OverviewPage=P.OverviewPage",
                    limit=1
                )
                
                if not player_details:
                    print(f"  No details found for {player_id}")
                    continue
                
                player = player_details[0]
                tournament_data = player_info['tournament_data']
                
                # Get team information
                team_details = {}
                if tournament_data.get('Team'):
                    time.sleep(1)
                    team_response = site.cargo_client.query(
                        tables="Teams",
                        fields="Name, Image",
                        where=f"OverviewPage = '{tournament_data.get('Team', '')}'",
                        limit=1
                    )
                    if team_response:
                        team_details = team_response[0]
                
                # Get KDA statistics
                print(f"  Getting KDA stats...")
                kda_stats = get_kda_stats(site, player_id, 2025)
                
                # Get number of teams/clubs
                clubs_count = 0
                try:
                    time.sleep(1)
                    tenures_response = site.cargo_client.query(
                        tables="PlayerRedirects=PR, Tenures=Te",
                        fields="Te.Team",
                        where=f"PR.AllName = '{player_id}'",
                        join_on="PR.OverviewPage=Te.Player",
                        limit=50
                    )
                    
                    if tenures_response:
                        unique_teams = set(t.get('Team', '') for t in tenures_response if t.get('Team'))
                        clubs_count = len(unique_teams)
                        
                except Exception as e:
                    print(f"    Error getting tenure data: {e}")
                
                # Get most played champion
                most_played_champion = ""
                try:
                    time.sleep(1)
                    champion_response = site.cargo_client.query(
                        tables="ScoreboardPlayers",
                        fields="Champion, COUNT(*) as ChampionCount",
                        where=f"Link = '{player_id}' AND DateTime_UTC LIKE '2025%'",
                        group_by="Champion",
                        order_by="ChampionCount DESC",
                        limit=1
                    )
                    
                    if champion_response:
                        most_played_champion = champion_response[0].get("Champion", "")
                        
                except Exception as e:
                    print(f"    Error getting champion data: {e}")
                
                # Get player image URL
                player_image_url = ""
                if player.get('Image'):
                    try:
                        image_file = player.get('Image', '').split('/')[-1]
                        player_image_url = f"https://lol.fandom.com/wiki/Special:FilePath/{image_file}"
                    except Exception as e:
                        print(f"    Error processing player image: {e}")
                
                team_image_url = ""
                if team_details.get('Image'):
                    team_image_file = team_details.get('Image', '').split('/')[-1]
                    team_image_url = f"https://lol.fandom.com/wiki/Special:FilePath/{team_image_file}"
                
                # Build improved player data
                player_data = {
                    "player_username": player_id,  # The gaming username/handle
                    "player_name": player.get("Name", "") or player_id,  # Real name
                    "player_media_url": player_image_url,
                    "player_team": team_details.get("Name", "") or tournament_data.get('Team', ''),
                    "player_team_media_url": team_image_url,
                    "player_league": "LEC",
                    "number_of_clubs": clubs_count,
                    "player_country": player.get("Country", ""),
                    "player_country_continent": get_continent_from_country(player.get("Country", "")),
                    "player_role": player.get("Role", "") or tournament_data.get('Role', ''),
                    "player_most_played_champion": most_played_champion,
                    "player_age": calculate_age(player.get("Birthdate", "")),
                    # KDA Stats
                    "avg_kills": kda_stats["avg_kills"],
                    "avg_deaths": kda_stats["avg_deaths"],
                    "avg_assists": kda_stats["avg_assists"],
                    "kda_ratio": kda_stats["kda_ratio"],
                    "games_played": kda_stats["games_played"]
                }
                
                processed_players.append(player_data)
                print(f"  ✓ Processed {player_data['player_name']} ({player_data['player_username']}) - KDA: {kda_stats['kda_ratio']}")
                
                # Fixed delay between players
                print(f"  Waiting 1s before next player...")
                time.sleep(1)
                
            except Exception as e:
                print(f"  Error processing {player_id}: {e}")
                continue
        
        return processed_players
        
    except Exception as e:
        print(f"Error in main query: {e}")
        import traceback
        traceback.print_exc()
        return []

def main():
    """Main function to execute the improved script"""
    print("Starting LEC 2025 Spring players data collection...")
    print("Features: Username, Real Name, Fixed Media URLs, KDA Stats, No Coaches\n")
    
    players_data = fetch_lec_players_improved()
    
    if not players_data:
        print("No player data found or error occurred.")
        return
    
    # Save to JSON file
    output_file = "players.json"
    
    try:
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(players_data, f, indent=2, ensure_ascii=False)
        
        print(f"\n🎉 Successfully saved {len(players_data)} LEC players data to {output_file}")
        
        # Print summary
        print(f"\nSummary:")
        print(f"Total players: {len(players_data)}")
        print(f"Players with age data: {len([p for p in players_data if p['player_age'] is not None])}")
        print(f"Players with KDA data: {len([p for p in players_data if p['games_played'] > 0])}")
        print(f"Players with champion data: {len([p for p in players_data if p['player_most_played_champion']])}")
        print(f"Players with team media: {len([p for p in players_data if p['player_team_media_url']])}")
        
        # Show sample of first few players
        print(f"\nSample players:")
        for player in players_data[:3]:
            print(f"  - {player['player_username']} ({player['player_name']}) - {player['player_role']} for {player['player_team']}")
            print(f"    KDA: {player['avg_kills']}/{player['avg_deaths']}/{player['avg_assists']} = {player['kda_ratio']} over {player['games_played']} games")
        
    except Exception as e:
        print(f"Error saving data to file: {e}")

if __name__ == "__main__":
    main()