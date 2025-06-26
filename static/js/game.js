// Main game logic for Prodle
class GameManager {
    constructor() {
        this.sessionId = '';
        this.currentPlayer = 1;
        this.totalPlayers = 20;
        this.score = 0;
        this.guessCount = 0;
        this.maxGuesses = 6;
        this.isGameActive = false;
        
        // DOM elements
        this.guessInput = document.getElementById('guess-input');
        this.guessButton = document.getElementById('guess-button');
        this.scoreElement = document.getElementById('score');
        this.playerCounterElement = document.getElementById('player-counter');
        this.guessesHistoryElement = document.getElementById('guesses-history');
        this.autocompleteList = document.getElementById('autocomplete-list');
        
        // Autocomplete
        this.selectedIndex = -1;
        this.autocompleteResults = [];
        
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
            this.guessInput.addEventListener('blur', () => this.hideAutocomplete());
        }

        // Guess button
        if (this.guessButton) {
            this.guessButton.addEventListener('click', () => this.makeGuess());
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
            item.addEventListener('click', () => this.selectAutocompleteItem(index));
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
        }, 150); // Small delay to allow click events
    }

    /**
     * Navigate autocomplete with arrow keys
     */
    navigateAutocomplete(direction) {
        if (this.autocompleteResults.length === 0) return;

        // Remove previous selection
        const items = this.autocompleteList.querySelectorAll('.autocomplete-item');
        if (this.selectedIndex >= 0 && items[this.selectedIndex]) {
            items[this.selectedIndex].classList.remove('selected');
        }

        // Update selection
        this.selectedIndex += direction;
        
        if (this.selectedIndex < 0) {
            this.selectedIndex = this.autocompleteResults.length - 1;
        } else if (this.selectedIndex >= this.autocompleteResults.length) {
            this.selectedIndex = 0;
        }

        // Add new selection
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
     * Make a guess
     */
    async makeGuess() {
        if (!this.isGameActive) return;

        const playerName = this.guessInput.value.trim();
        if (!playerName) {
            alert('Veuillez entrer le nom d\'un joueur');
            return;
        }

        this.guessButton.disabled = true;
        this.guessInput.disabled = true;

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

        this.guessButton.disabled = false;
        this.guessInput.disabled = false;
        this.guessInput.focus();
    }

    /**
     * Handle guess result
     */
    handleGuessResult(result) {
        // Update score
        this.score = result.score;
        this.updateScoreDisplay();

        // Add guess to history
        this.addGuessToHistory(result);

        // Clear input
        this.guessInput.value = '';
        this.hideAutocomplete();

        // Check if correct or should move to next player
        if (result.correct || result.nextPlayer) {
            this.moveToNextPlayer();
        }

        // Check if game is over
        if (result.gameOver) {
            this.handleGameOver();
        }

        this.guessCount++;
    }

    /**
     * Add guess result to history display
     */
    addGuessToHistory(result) {
        const guessDiv = document.createElement('div');
        guessDiv.className = 'guess-result';
        
        let content = `<strong>${result.comparison.guessed_player.player_username}</strong><br>`;
        
        if (result.correct) {
            content += '<span style="color: #32CD32;">✓ Correct!</span>';
        } else {
            content += '<span style="color: #FF4444;">✗ Incorrect</span>';
        }

        guessDiv.innerHTML = content;
        this.guessesHistoryElement.appendChild(guessDiv);
        
        // Scroll to bottom
        this.guessesHistoryElement.scrollTop = this.guessesHistoryElement.scrollHeight;
    }

    /**
     * Move to next player
     */
    moveToNextPlayer() {
        this.currentPlayer++;
        this.guessCount = 0;
        
        // Reset timer for next player
        window.timerManager.reset();
        window.timerManager.start(
            () => this.handleTimeUp(),
            (timeLeft) => this.handleTimerTick(timeLeft)
        );

        // Update player counter
        this.updatePlayerCounter();

        // Clear guess history for next player
        this.guessesHistoryElement.innerHTML = '';

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
     * Handle game over
     */
    handleGameOver() {
        this.isGameActive = false;
        window.timerManager.stop();
        
        console.log('Game over! Final score:', this.score);
        
        // Show game over dialog
        const username = prompt(`Jeu terminé! Score final: ${this.score}\nEntrez votre nom pour le classement:`);
        
        if (username && username.trim()) {
            this.submitScore(username.trim());
        } else {
            // Redirect to home after a delay
            setTimeout(() => {
                window.location.href = '/';
            }, 3000);
        }
    }

    /**
     * Submit final score
     */
    async submitScore(username) {
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
                alert('Score enregistré avec succès!');
            } else {
                alert('Erreur lors de l\'enregistrement: ' + (data.message || 'Erreur inconnue'));
            }
        } catch (error) {
            console.error('Error submitting score:', error);
            alert('Erreur de connexion lors de l\'enregistrement');
        }

        // Redirect to home
        setTimeout(() => {
            window.location.href = '/';
        }, 2000);
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
        // For now, we'll leave it empty since the endpoint isn't created yet
        console.log('Leaderboard loading not implemented yet');
    }
}

// Global game manager instance
window.gameManager = new GameManager();

// Global function for guess button (called from HTML)
function makeGuess() {
    if (window.gameManager) {
        window.gameManager.makeGuess();
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    console.log('Game system initialized');
});