#!/usr/bin/env python3
"""
Script to download team images from the players.json file and save them locally.
This avoids CORS and 404 issues with external image URLs.
"""

import json
import requests
import os
import re
from urllib.parse import urlparse
import time

def clean_filename(filename):
    """Clean filename to be safe for filesystem"""
    # Remove or replace unsafe characters
    filename = re.sub(r'[<>:"/\\|?*]', '', filename)
    # Replace spaces with underscores
    filename = filename.replace(' ', '_')
    # Remove multiple consecutive underscores
    filename = re.sub(r'_+', '_', filename)
    return filename

def get_team_image_filename(team_name, url):
    """Generate a safe filename for the team image"""
    clean_team = clean_filename(team_name)
    
    # Try to get file extension from URL
    parsed_url = urlparse(url)
    path = parsed_url.path
    if '.' in path:
        ext = path.split('.')[-1].lower()
        # Only use common image extensions
        if ext in ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg']:
            return f"{clean_team}.{ext}"
    
    # Default to .png if no extension found
    return f"{clean_team}.png"

def download_image(url, filepath, team_name):
    """Download an image from URL to filepath"""
    try:
        print(f"Downloading {team_name}: {url}")
        
        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
            'Accept': 'image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8',
        }
        
        response = requests.get(url, headers=headers, timeout=30, allow_redirects=True)
        
        if response.status_code == 200:
            with open(filepath, 'wb') as f:
                f.write(response.content)
            print(f"  ✓ Downloaded successfully")
            return True
        else:
            print(f"  ✗ Failed: HTTP {response.status_code}")
            return False
            
    except Exception as e:
        print(f"  ✗ Error: {str(e)}")
        return False

def main():
    # Load players data
    with open('players.json', 'r') as f:
        players_data = json.load(f)
    
    # Create static directory if it doesn't exist
    static_dir = 'static'
    team_images_dir = os.path.join(static_dir, 'team_images')
    os.makedirs(team_images_dir, exist_ok=True)
    
    # Collect unique team URLs
    team_urls = {}
    for player in players_data:
        team_name = player.get('player_team', '')
        team_url = player.get('player_team_media_url', '')
        
        if team_name and team_url and team_url.strip():
            if team_name not in team_urls:
                team_urls[team_name] = team_url.strip()
    
    print(f"Found {len(team_urls)} unique teams with image URLs:")
    for team, url in team_urls.items():
        print(f"  {team}: {url}")
    
    print("\nStarting downloads...")
    
    successful_downloads = {}
    failed_downloads = {}
    
    for team_name, url in team_urls.items():
        filename = get_team_image_filename(team_name, url)
        filepath = os.path.join(team_images_dir, filename)
        
        # Skip if file already exists
        if os.path.exists(filepath):
            print(f"Skipping {team_name}: file already exists")
            successful_downloads[team_name] = filename
            continue
        
        # Download the image
        if download_image(url, filepath, team_name):
            successful_downloads[team_name] = filename
        else:
            failed_downloads[team_name] = url
        
        # Add a small delay to be respectful to servers
        time.sleep(0.5)
    
    print(f"\n=== Download Summary ===")
    print(f"Successful: {len(successful_downloads)}")
    print(f"Failed: {len(failed_downloads)}")
    
    if successful_downloads:
        print(f"\n✓ Successfully downloaded:")
        for team, filename in successful_downloads.items():
            print(f"  {team} -> {filename}")
    
    if failed_downloads:
        print(f"\n✗ Failed downloads:")
        for team, url in failed_downloads.items():
            print(f"  {team}: {url}")
    
    # Create a mapping file for the application to use
    mapping_file = os.path.join(static_dir, 'team_image_mapping.json')
    with open(mapping_file, 'w') as f:
        json.dump(successful_downloads, f, indent=2)
    
    print(f"\nCreated team image mapping file: {mapping_file}")

if __name__ == "__main__":
    main()