// Complete game flow implementation for Prodle
class GameManager {
    constructor() {
        this.sessionId = '';
        this.currentPlayer = 1;
        this.totalPlayers = 20;
        this.score = 0;
        this.guessCount = 0;
        this.maxGuesses = 6;
        this.isGameActive = false;
        this.currentTargetPlayer = null;
        this.playersFound = 0;
        
        // DOM elements
        this.guessInput = document.getElementById('guess-input');
        this.guessButton = document.getElementById('guess-button');
        this.scoreElement = document.getElementById('score');
        this.playerCounterElement = document.getElementById('player-counter');
        this.guessHistoryElement = document.getElementById('guess-list');
        this.autocompleteList = document.getElementById('autocomplete-list');
        this.currentPlayerDisplay = document.getElementById('current-player-display');
        this.currentPlayerGrid = document.getElementById('current-player-grid');
        
        // Overlay elements
        this.successOverlay = document.getElementById('success-overlay');
        this.endGameOverlay = document.getElementById('end-game-overlay');
        this.loadingOverlay = document.getElementById('loading-overlay');
        this.finalScoreElement = document.getElementById('final-score');
        this.playersCompletedElement = document.getElementById('players-completed');
        this.usernameInput = document.getElementById('username-input');
        this.scoreForm = document.getElementById('score-form');
        this.scoreSubmitted = document.getElementById('score-submitted');
        this.submitScoreBtn = document.getElementById('submit-score-btn');
        
        // Autocomplete
        this.selectedIndex = -1;
        this.autocompleteResults = [];
        
        // Country flag mapping
        this.countryFlags = {
            'South Korea': '🇰🇷',
            'Denmark': '🇩🇰',
            'Germany': '🇩🇪',
            'France': '🇫🇷',
            'Spain': '🇪🇸',
            'Poland': '🇵🇱',
            'Sweden': '🇸🇪',
            'Belgium': '🇧🇪',
            'Netherlands': '🇳🇱',
            'Czech Republic': '🇨🇿',
            'United States': '🇺🇸',
            'Canada': '🇨🇦',
            'Australia': '🇦🇺',
            'United Kingdom': '🇬🇧',
            'Ireland': '🇮🇪',
            'Norway': '🇳🇴',
            'Finland': '🇫🇮',
            'Italy': '🇮🇹',
            'Slovenia': '🇸🇮',
            'Bulgaria': '🇧🇬',
            'Turkey': '🇹🇷',
            'Greece': '🇬🇷',
            'China': '🇨🇳',
            'Japan': '🇯🇵',
            'Taiwan': '🇹🇼'
        };
        
        this.setupEventListeners();
    }

    /**
     * Initialize the game
     */
    initialize() {
        this.sessionId = document.getElementById('session-id').value;
        this.isGameActive = true;
        
        // Disable controls initially (enabled after countdown)
        this.guessInput.disabled = true;
        this.guessButton.disabled = true;
        
        // Setup timer callbacks
        window.timerManager.onTimeUp = () => this.handleTimeUp();
        window.timerManager.onTick = (timeLeft) => this.handleTimerTick(timeLeft);
        
        // Load initial leaderboard
        this.loadLeaderboard();
        
        console.log('Game initialized with session:', this.sessionId);
    }

    /**
     * Setup event listeners
     */
    setupEventListeners() {
        // Guess input events
        if (this.guessInput) {
            this.guessInput.addEventListener('input', (e) => this.handleInputChange(e));
            this.guessInput.addEventListener('keydown', (e) => this.handleKeyDown(e));
            this.guessInput.addEventListener('focus', () => {
                if (this.guessInput.value.length >= 2) {
                    this.showAutocomplete();
                }
            });
            this.guessInput.addEventListener('blur', () => this.hideAutocomplete());
        }

        // Guess button
        if (this.guessButton) {
            this.guessButton.addEventListener('click', () => this.makeGuess());
        }

        // Username input enter key
        if (this.usernameInput) {
            this.usernameInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    this.submitFinalScore();
                }
            });
        }
    }

    /**
     * Show loading overlay
     */
    showLoading(text = 'Chargement...') {
        if (this.loadingOverlay) {
            this.loadingOverlay.querySelector('.loading-text').textContent = text;
            this.loadingOverlay.classList.remove('hidden');
        }
    }

    /**
     * Hide loading overlay
     */
    hideLoading() {
        if (this.loadingOverlay) {
            this.loadingOverlay.classList.add('hidden');
        }
    }

    /**
     * Handle input change for autocomplete
     */
    async handleInputChange(event) {
        const query = event.target.value.trim();
        
        if (query.length >= 2) {
            await this.fetchAutocomplete(query);
        } else {
            this.hideAutocomplete();
        }
    }

    /**
     * Handle keyboard navigation
     */
    handleKeyDown(event) {
        if (event.key === 'Enter') {
            event.preventDefault();
            if (this.selectedIndex >= 0 && this.autocompleteResults.length > 0) {
                this.selectAutocompleteItem(this.selectedIndex);
            } else {
                this.makeGuess();
            }
        } else if (event.key === 'ArrowDown') {
            event.preventDefault();
            this.navigateAutocomplete(1);
        } else if (event.key === 'ArrowUp') {
            event.preventDefault();
            this.navigateAutocomplete(-1);
        } else if (event.key === 'Escape') {
            this.hideAutocomplete();
        }
    }

    /**
     * Fetch autocomplete suggestions
     */
    async fetchAutocomplete(query) {
        try {
            const response = await fetch(`/api/autocomplete?query=${encodeURIComponent(query)}`);
            const data = await response.json();
            
            this.autocompleteResults = data.players || [];
            this.selectedIndex = -1;
            this.showAutocomplete();
        } catch (error) {
            console.error('Error fetching autocomplete:', error);
            this.hideAutocomplete();
        }
    }

    /**
     * Show autocomplete dropdown
     */
    showAutocomplete() {
        if (this.autocompleteResults.length === 0) {
            this.hideAutocomplete();
            return;
        }

        this.autocompleteList.innerHTML = '';
        
        this.autocompleteResults.forEach((player, index) => {
            const item = document.createElement('div');
            item.className = 'autocomplete-item';
            item.textContent = player;
            item.addEventListener('mousedown', (e) => {
                e.preventDefault();
                this.selectAutocompleteItem(index);
            });
            this.autocompleteList.appendChild(item);
        });

        this.autocompleteList.classList.remove('hidden');
    }

    /**
     * Hide autocomplete dropdown
     */
    hideAutocomplete() {
        setTimeout(() => {
            this.autocompleteList.classList.add('hidden');
        }, 150);
    }

    /**
     * Navigate autocomplete with arrow keys
     */
    navigateAutocomplete(direction) {
        if (this.autocompleteResults.length === 0) return;

        const items = this.autocompleteList.querySelectorAll('.autocomplete-item');
        if (this.selectedIndex >= 0 && items[this.selectedIndex]) {
            items[this.selectedIndex].classList.remove('selected');
        }

        this.selectedIndex += direction;
        
        if (this.selectedIndex < 0) {
            this.selectedIndex = this.autocompleteResults.length - 1;
        } else if (this.selectedIndex >= this.autocompleteResults.length) {
            this.selectedIndex = 0;
        }

        if (items[this.selectedIndex]) {
            items[this.selectedIndex].classList.add('selected');
            items[this.selectedIndex].scrollIntoView({ block: 'nearest' });
        }
    }

    /**
     * Select an autocomplete item
     */
    selectAutocompleteItem(index) {
        if (index >= 0 && index < this.autocompleteResults.length) {
            this.guessInput.value = this.autocompleteResults[index];
            this.hideAutocomplete();
            this.guessInput.focus();
        }
    }

    /**
     * Make a guess - Task 25: Implement guess submission flow
     */
    async makeGuess() {
        if (!this.isGameActive) return;

        const playerName = this.guessInput.value.trim();
        if (!playerName) {
            alert('Veuillez entrer le nom d\'un joueur');
            return;
        }

        // Show loading state
        this.guessButton.disabled = true;
        this.guessInput.disabled = true;
        this.guessButton.textContent = 'Traitement...';

        try {
            const response = await fetch('/api/guess', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    sessionId: this.sessionId,
                    playerName: playerName
                })
            });

            const data = await response.json();
            
            if (data.success) {
                this.handleGuessResult(data);
            } else {
                alert('Erreur: ' + (data.message || 'Erreur inconnue'));
            }
        } catch (error) {
            console.error('Error making guess:', error);
            alert('Erreur de connexion au serveur');
        }

        // Reset button state
        this.guessButton.disabled = false;
        this.guessInput.disabled = false;
        this.guessButton.textContent = 'Deviner';
        this.guessInput.focus();
    }

    /**
     * Handle guess result with animations
     */
    handleGuessResult(result) {
        // Update score with animation
        this.score = result.score;
        this.updateScoreDisplay();

        // Add guess to history with detailed display and animations
        this.addGuessToHistory(result);

        // Clear input
        this.guessInput.value = '';
        this.hideAutocomplete();

        // Check if correct - Task 26: Handle correct guess flow
        if (result.correct) {
            this.handleCorrectGuess(result);
        } else if (result.nextPlayer) {
            // Max guesses reached, move to next player
            setTimeout(() => {
                this.moveToNextPlayer();
            }, 1000);
        }

        // Check if game is over
        if (result.gameOver) {
            setTimeout(() => {
                this.handleGameOver();
            }, result.correct ? 4000 : 2000);
        }

        this.guessCount++;
    }

    /**
     * Task 26: Handle correct guess flow
     */
    handleCorrectGuess(result) {
        this.playersFound++;
        
        // Show success message "Bravo!" for 1 second
        this.showSuccessMessage();
        
        // Show current player reveal
        setTimeout(() => {
            this.showCurrentPlayerReveal(result.comparison.guessed_player);
        }, 1000);
        
        // Fade out and move to next player after 3 seconds total
        setTimeout(() => {
            this.fadeOutGameState();
            setTimeout(() => {
                this.moveToNextPlayer();
            }, 500);
        }, 3000);
    }

    /**
     * Show success message overlay
     */
    showSuccessMessage() {
        if (this.successOverlay) {
            this.successOverlay.classList.remove('hidden');
            
            // Hide after 1 second
            setTimeout(() => {
                this.successOverlay.classList.add('hidden');
            }, 1000);
        }
    }

    /**
     * Fade out current game state
     */
    fadeOutGameState() {
        const mainGame = document.querySelector('.main-game');
        if (mainGame) {
            mainGame.style.transition = 'opacity 0.5s ease-out';
            mainGame.style.opacity = '0.3';
            
            // Reset opacity after transition
            setTimeout(() => {
                mainGame.style.opacity = '1';
            }, 1000);
        }
    }

    /**
     * Create player attribute card
     */
    createPlayerAttributeCard(label, value, comparisonResult, attributeKey, targetValue = null) {
        const card = document.createElement('div');
        card.className = `player-attribute ${comparisonResult}`;
        
        const labelDiv = document.createElement('div');
        labelDiv.className = 'attribute-label';
        labelDiv.textContent = label;
        
        const valueDiv = document.createElement('div');
        valueDiv.className = 'attribute-value';
        
        // Handle special display cases
        if (attributeKey === 'team') {
            const teamImg = document.createElement('img');
            teamImg.className = 'team-image';
            teamImg.src = `/assets/teams/${value.replace(/\s+/g, '_')}.png`;
            teamImg.alt = value;
            teamImg.onerror = () => {
                teamImg.style.display = 'none';
            };
            valueDiv.appendChild(teamImg);
            
            const teamText = document.createElement('span');
            teamText.textContent = value;
            valueDiv.appendChild(teamText);
        } else if (attributeKey === 'country') {
            const flag = this.countryFlags[value] || '🏳️';
            const flagSpan = document.createElement('span');
            flagSpan.className = 'country-flag';
            flagSpan.textContent = flag;
            valueDiv.appendChild(flagSpan);
            
            const countryText = document.createElement('span');
            countryText.textContent = value;
            valueDiv.appendChild(countryText);
        } else if (attributeKey === 'champion') {
            const champImg = document.createElement('img');
            champImg.className = 'champion-image';
            champImg.src = this.getChampionImageUrl(value);
            champImg.alt = value;
            champImg.onerror = () => {
                champImg.style.display = 'none';
            };
            valueDiv.appendChild(champImg);
            
            const champText = document.createElement('span');
            champText.textContent = value;
            valueDiv.appendChild(champText);
        } else {
            valueDiv.textContent = value;
        }
        
        // Add arrow indicators for numerical values
        if ((attributeKey === 'age' || attributeKey === 'kda') && comparisonResult === 'incorrect' && targetValue !== null) {
            const arrow = document.createElement('span');
            arrow.className = 'arrow-indicator';
            
            const numValue = parseFloat(value);
            const numTarget = parseFloat(targetValue);
            
            if (numValue > numTarget) {
                arrow.textContent = '↓';
                arrow.classList.add('arrow-down');
            } else if (numValue < numTarget) {
                arrow.textContent = '↑';
                arrow.classList.add('arrow-up');
            }
            
            card.appendChild(arrow);
        }
        
        card.appendChild(labelDiv);
        card.appendChild(valueDiv);
        
        return card;
    }

    /**
     * Get champion image URL
     */
    getChampionImageUrl(championName) {
        const nameMap = {
            "Kai'Sa": "Kaisa",
            "Wukong": "MonkeyKing",
            "Renata Glasc": "Renata"
        };
        
        const mappedName = nameMap[championName] || championName;
        return `https://ddragon.leagueoflegends.com/cdn/img/champion/centered/${mappedName}_0.jpg`;
    }

    /**
     * Get comparison result class name
     */
    getComparisonClass(comparisonResult) {
        switch (comparisonResult) {
            case 'exact':
                return 'correct';
            case 'partial':
                return 'partial';
            case 'higher':
            case 'lower':
            case 'wrong':
                return 'incorrect';
            default:
                return 'incorrect';
        }
    }

    /**
     * Add guess result to history with animations
     */
    addGuessToHistory(result) {
        const guessDiv = document.createElement('div');
        guessDiv.className = result.correct ? 'guess-result correct-guess' : 'guess-result';
        
        // Add fade-in animation
        guessDiv.style.opacity = '0';
        guessDiv.style.transform = 'translateY(20px)';
        
        // Header with player name and status
        const header = document.createElement('div');
        header.className = 'guess-result-header';
        
        const playerName = document.createElement('div');
        playerName.className = 'guess-player-name';
        playerName.textContent = result.comparison.guessed_player.player_username;
        
        const status = document.createElement('div');
        status.className = `guess-result-status ${result.correct ? 'correct' : 'incorrect'}`;
        status.textContent = result.correct ? '✓ Correct!' : '✗ Incorrect';
        
        header.appendChild(playerName);
        header.appendChild(status);
        
        // Attributes grid
        const attributesGrid = document.createElement('div');
        attributesGrid.className = 'guess-attributes';
        
        const player = result.comparison.guessed_player;
        const comparisons = result.comparison.comparisons;
        
        const attributes = [
            { key: 'team', label: 'Équipe', value: player.player_team },
            { key: 'league', label: 'Ligue', value: player.player_league },
            { key: 'role', label: 'Rôle', value: player.player_role },
            { key: 'country', label: 'Pays', value: player.player_country },
            { key: 'age', label: 'Âge', value: player.player_age },
            { key: 'clubs', label: 'Clubs', value: player.number_of_clubs },
            { key: 'kda', label: 'KDA', value: player.kda_ratio.toFixed(2) },
            { key: 'champion', label: 'Champion', value: player.player_most_played_champion }
        ];
        
        attributes.forEach(attr => {
            const comparisonResult = comparisons[attr.key] || 'wrong';
            const attrDiv = document.createElement('div');
            attrDiv.className = `guess-attribute ${this.getComparisonClass(comparisonResult)}`;
            attrDiv.textContent = `${attr.label}: ${attr.value}`;
            attributesGrid.appendChild(attrDiv);
        });
        
        guessDiv.appendChild(header);
        guessDiv.appendChild(attributesGrid);
        
        this.guessHistoryElement.appendChild(guessDiv);
        
        // Animate in
        setTimeout(() => {
            guessDiv.style.transition = 'all 0.5s ease-out';
            guessDiv.style.opacity = '1';
            guessDiv.style.transform = 'translateY(0)';
        }, 50);
        
        // Scroll to bottom
        this.guessHistoryElement.scrollTop = this.guessHistoryElement.scrollHeight;
    }

    /**
     * Show current player reveal with full details
     */
    showCurrentPlayerReveal(player) {
        this.currentPlayerDisplay.classList.remove('hidden');
        this.currentPlayerDisplay.classList.add('revealed');
        
        // Clear and populate player grid
        this.currentPlayerGrid.innerHTML = '';
        
        const attributes = [
            { key: 'name', label: 'Joueur', value: player.player_username },
            { key: 'team', label: 'Équipe', value: player.player_team },
            { key: 'league', label: 'Ligue', value: player.player_league },
            { key: 'role', label: 'Rôle', value: player.player_role },
            { key: 'country', label: 'Pays', value: player.player_country },
            { key: 'age', label: 'Âge', value: player.player_age.toString() },
            { key: 'clubs', label: 'Clubs', value: player.number_of_clubs.toString() },
            { key: 'kda', label: 'KDA', value: player.kda_ratio.toFixed(2) },
            { key: 'champion', label: 'Champion', value: player.player_most_played_champion }
        ];
        
        attributes.forEach(attr => {
            const card = this.createPlayerAttributeCard(attr.label, attr.value, 'correct', attr.key);
            this.currentPlayerGrid.appendChild(card);
        });
    }

    /**
     * Move to next player - Load next player and reset guess history
     */
    moveToNextPlayer() {
        this.currentPlayer++;
        this.guessCount = 0;
        
        // Hide current player display
        this.currentPlayerDisplay.classList.add('hidden');
        this.currentPlayerDisplay.classList.remove('revealed');
        
        // Reset guess history for next player
        this.guessHistoryElement.innerHTML = '';
        
        // Reset timer for next player and continue timer
        window.timerManager.reset();
        window.timerManager.start(
            () => this.handleTimeUp(),
            (timeLeft) => this.handleTimerTick(timeLeft)
        );

        // Update player counter
        this.updatePlayerCounter();

        console.log(`Moved to player ${this.currentPlayer}/${this.totalPlayers}`);
    }

    /**
     * Handle time up for current player
     */
    handleTimeUp() {
        console.log('Time up for current player');
        
        if (this.currentPlayer < this.totalPlayers) {
            this.moveToNextPlayer();
        } else {
            this.handleGameOver();
        }
    }

    /**
     * Handle timer tick
     */
    handleTimerTick(timeLeft) {
        // Could add additional UI updates here if needed
    }

    /**
     * Task 27: Create end game flow
     */
    handleGameOver() {
        this.isGameActive = false;
        window.timerManager.stop();
        
        console.log('Game over! Final score:', this.score);
        
        // Show final score and username input form
        this.showEndGameOverlay();
    }

    /**
     * Show end game overlay with final score and username input
     */
    showEndGameOverlay() {
        if (this.finalScoreElement) {
            this.finalScoreElement.textContent = `Score Final: ${this.score}`;
        }
        
        if (this.playersCompletedElement) {
            this.playersCompletedElement.textContent = `Joueurs Trouvés: ${this.playersFound}/${this.totalPlayers}`;
        }
        
        // Show score form, hide submitted message
        this.scoreForm.classList.remove('hidden');
        this.scoreSubmitted.classList.add('hidden');
        
        // Focus username input
        setTimeout(() => {
            if (this.usernameInput) {
                this.usernameInput.focus();
            }
        }, 500);
        
        this.endGameOverlay.classList.remove('hidden');
    }

    /**
     * Task 28: Implement score submission
     */
    async submitFinalScore() {
        const username = this.usernameInput.value.trim();
        
        if (!username) {
            alert('Veuillez entrer votre nom');
            this.usernameInput.focus();
            return;
        }

        if (username.length > 50) {
            alert('Le nom ne peut pas dépasser 50 caractères');
            return;
        }

        // Disable submit button and show loading
        this.submitScoreBtn.disabled = true;
        this.submitScoreBtn.textContent = 'Enregistrement...';

        try {
            const response = await fetch('/api/submit-score', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    sessionId: this.sessionId,
                    username: username
                })
            });

            const data = await response.json();
            
            if (data.success) {
                // Show confirmation message and update leaderboard
                this.showScoreSubmitted();
                this.loadLeaderboard(); // Refresh leaderboard
            } else {
                alert('Erreur lors de l\'enregistrement: ' + (data.message || 'Erreur inconnue'));
                this.submitScoreBtn.disabled = false;
                this.submitScoreBtn.textContent = 'Enregistrer Score';
            }
        } catch (error) {
            console.error('Error submitting score:', error);
            alert('Erreur de connexion lors de l\'enregistrement');
            this.submitScoreBtn.disabled = false;
            this.submitScoreBtn.textContent = 'Enregistrer Score';
        }
    }

    /**
     * Show score submitted confirmation
     */
    showScoreSubmitted() {
        this.scoreForm.classList.add('hidden');
        this.scoreSubmitted.classList.remove('hidden');
    }

    /**
     * Task 29: Build restart functionality
     */
    async restartGame() {
        this.showLoading('Création d\'une nouvelle partie...');
        
        try {
            // Create new session
            const response = await fetch('/api/start-game', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                }
            });
            
            const data = await response.json();
            
            if (data.success) {
                // Store new session ID and reset all game state
                sessionStorage.setItem('sessionId', data.sessionId);
                this.resetGameState();
                
                // Hide all overlays
                this.hideLoading();
                this.endGameOverlay.classList.add('hidden');
                
                // Show countdown again and start fresh game
                window.countdownManager.start(() => {
                    console.log('Countdown finished, starting fresh game...');
                    
                    if (window.timerManager) {
                        window.timerManager.start(
                            () => this.handleTimeUp(),
                            (timeLeft) => this.handleTimerTick(timeLeft)
                        );
                    }
                    
                    // Re-enable game controls
                    this.guessInput.disabled = false;
                    this.guessButton.disabled = false;
                    this.guessInput.focus();
                    
                    // Re-initialize game
                    this.initialize();
                });
            } else {
                alert('Erreur lors du démarrage du jeu: ' + (data.message || 'Erreur inconnue'));
                this.hideLoading();
            }
        } catch (error) {
            console.error('Error restarting game:', error);
            alert('Erreur de connexion au serveur');
            this.hideLoading();
        }
    }

    /**
     * Reset all game state for restart
     */
    resetGameState() {
        this.sessionId = sessionStorage.getItem('sessionId');
        this.currentPlayer = 1;
        this.score = 0;
        this.guessCount = 0;
        this.playersFound = 0;
        this.isGameActive = true;
        
        // Update displays
        this.updateScoreDisplay();
        this.updatePlayerCounter();
        
        // Clear guess history
        this.guessHistoryElement.innerHTML = '';
        
        // Hide player display
        this.currentPlayerDisplay.classList.add('hidden');
        this.currentPlayerDisplay.classList.remove('revealed');
        
        // Clear input
        this.guessInput.value = '';
        this.usernameInput.value = '';
        
        // Reset submit button
        this.submitScoreBtn.disabled = false;
        this.submitScoreBtn.textContent = 'Enregistrer Score';
        
        // Update session ID input
        document.getElementById('session-id').value = this.sessionId;
        
        console.log('Game state reset for restart');
    }

    /**
     * Update score display
     */
    updateScoreDisplay() {
        if (this.scoreElement) {
            this.scoreElement.textContent = `Score: ${this.score}`;
        }
    }

    /**
     * Update player counter display
     */
    updatePlayerCounter() {
        if (this.playerCounterElement) {
            this.playerCounterElement.textContent = `Joueur ${this.currentPlayer}/${this.totalPlayers}`;
        }
    }

    /**
     * Load leaderboard for sidebar
     */
    async loadLeaderboard() {
        // This would call a leaderboard API endpoint when implemented
        console.log('Leaderboard loading not implemented yet');
    }
}

// Global game manager instance
window.gameManager = new GameManager();

// Global functions for HTML onclick handlers
function makeGuess() {
    if (window.gameManager) {
        window.gameManager.makeGuess();
    }
}

function submitFinalScore() {
    if (window.gameManager) {
        window.gameManager.submitFinalScore();
    }
}

function restartGame() {
    if (window.gameManager) {
        window.gameManager.restartGame();
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    console.log('Complete game flow system initialized');
});