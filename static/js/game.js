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
        this.isTransitioning = false;
        
        // DOM elements
        this.guessInput = document.getElementById('guess-input');
        this.guessButton = document.getElementById('guess-button');
        this.scoreElement = document.getElementById('score');
        this.playerCounterElement = document.getElementById('player-counter');
        this.guessRowsElement = document.getElementById('guess-rows');
        this.autocompleteList = document.getElementById('autocomplete-list');
        
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
        this.autocompleteDebounceTimer = null;
        this.debounceDelay = 10; // milliseconds
        
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
     * Initialize the game for first time
     */
    initialize() {
        // Get session ID from sessionStorage (set when starting game from index page)
        this.sessionId = sessionStorage.getItem('sessionId') || document.getElementById('session-id').value;
        
        if (!this.sessionId) {
            console.error('No session ID found! Redirecting to home page.');
            window.location.href = '/';
            return;
        }
        
        // Update the hidden input for other functions that might need it
        document.getElementById('session-id').value = this.sessionId;
        
        this.isGameActive = true;
        
        // Setup timer callbacks
        window.timerManager.onTimeUp = () => this.handleTimeUp();
        window.timerManager.onTick = (timeLeft) => this.handleTimerTick(timeLeft);
        
        // Load initial leaderboard
        this.loadLeaderboard();
        
        console.log('Game initialized');
    }

    /**
     * Setup initial game state (called before countdown)
     */
    setupInitialState() {
        // Get session ID from sessionStorage
        this.sessionId = sessionStorage.getItem('sessionId') || document.getElementById('session-id').value;
        
        if (!this.sessionId) {
            console.error('No session ID found during setup! Redirecting to home page.');
            window.location.href = '/';
            return;
        }
        
        // Update the hidden input
        document.getElementById('session-id').value = this.sessionId;
        
        this.isGameActive = false; // Will be enabled after countdown
        
        // Disable controls initially (enabled after countdown)
        if (this.guessInput) this.guessInput.disabled = true;
        if (this.guessButton) this.guessButton.disabled = true;
        
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

        // Guess button (if exists)
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
     * Set loading state for input
     */
    setLoadingState(loading) {
        if (loading) {
            this.guessInput.disabled = true;
            this.guessInput.classList.add('loading');
            if (this.guessButton) {
                this.guessButton.disabled = true;
                this.guessButton.innerHTML = '<span class="spinner"></span> Traitement...';
                this.guessButton.classList.add('loading');
            }
        } else {
            this.guessInput.disabled = false;
            this.guessInput.classList.remove('loading');
            if (this.guessButton) {
                this.guessButton.disabled = false;
                this.guessButton.textContent = 'Deviner';
                this.guessButton.classList.remove('loading');
            }
            this.guessInput.focus();
        }
    }

    /**
     * Set input disabled state during transitions
     */
    setInputDisabled(disabled) {
        if (this.guessInput) {
            this.guessInput.disabled = disabled;
            if (disabled) {
                this.guessInput.classList.add('transitioning');
            } else {
                this.guessInput.classList.remove('transitioning');
                this.guessInput.focus();
            }
        }
        if (this.guessButton) {
            this.guessButton.disabled = disabled;
        }
        // Hide autocomplete during transitions
        if (disabled) {
            this.hideAutocomplete();
        }
    }

    /**
     * Show user-friendly error message
     */
    showUserFriendlyError(message) {
        // Create temporary error element
        const errorEl = document.createElement('div');
        errorEl.className = 'error-message';
        errorEl.textContent = message;
        errorEl.style.cssText = `
            position: fixed;
            top: 20px;
            left: 50%;
            transform: translateX(-50%);
            background: rgba(220, 53, 69, 0.9);
            color: white;
            padding: 1rem 1.5rem;
            border-radius: 8px;
            z-index: 10000;
            font-weight: bold;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
            animation: slideDown 0.3s ease;
        `;

        document.body.appendChild(errorEl);

        // Remove after 4 seconds
        setTimeout(() => {
            errorEl.style.animation = 'slideUp 0.3s ease forwards';
            setTimeout(() => errorEl.remove(), 300);
        }, 4000);
    }

    /**
     * Handle API errors
     */
    handleApiError(message) {
        if (message.includes('Session not found') || message.includes('Session')) {
            this.showUserFriendlyError('Session expirée. Veuillez redémarrer le jeu.');
            setTimeout(() => {
                window.location.href = '/';
            }, 3000);
        } else {
            this.showUserFriendlyError(message);
        }
    }

    /**
     * Handle network errors
     */
    handleNetworkError(error) {
        if (error.name === 'TypeError' && error.message.includes('fetch')) {
            this.showUserFriendlyError('Problème de connexion. Vérifiez votre connexion internet.');
        } else if (error.message.includes('timeout')) {
            this.showUserFriendlyError('Délai d\'attente dépassé. Veuillez réessayer.');
        } else {
            this.showUserFriendlyError('Erreur de connexion au serveur');
        }
    }

    /**
     * Handle input change for autocomplete with debouncing
     */
    handleInputChange(event) {
        // Don't process input changes during transitions
        if (this.isTransitioning) return;
        
        const query = event.target.value.trim();
        
        // Clear existing timer
        if (this.autocompleteDebounceTimer) {
            clearTimeout(this.autocompleteDebounceTimer);
        }
        
        if (query.length >= 2) {
            // Add loading state to input
            this.guessInput.classList.add('loading');
            
            // Debounce the API call
            this.autocompleteDebounceTimer = setTimeout(async () => {
                await this.fetchAutocomplete(query);
                this.guessInput.classList.remove('loading');
            }, this.debounceDelay);
        } else {
            this.hideAutocomplete();
            this.guessInput.classList.remove('loading');
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
            } else if (this.autocompleteResults.length > 0) {
                // If no player is selected but there are autocomplete results, select the first one
                this.selectAutocompleteItem(0);
            } else {
                // No autocomplete results - show error message
                this.showUserFriendlyError('Ce joueur n\'existe pas');
                this.guessInput.focus();
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
     * Show autocomplete dropdown with highlighted text
     */
    showAutocomplete() {
        const query = this.guessInput.value.trim().toLowerCase();
        
        if (this.autocompleteResults.length === 0) {
            this.hideAutocomplete();
            return;
        }

        this.autocompleteList.innerHTML = '';
        
        // Limit to 10 results for performance
        const limitedResults = this.autocompleteResults.slice(0, 10);
        
        limitedResults.forEach((player, index) => {
            const item = document.createElement('div');
            item.className = 'autocomplete-item';
            
            // Highlight matching text
            const playerLower = player.toLowerCase();
            const queryIndex = playerLower.indexOf(query);
            
            if (queryIndex !== -1) {
                const before = player.substring(0, queryIndex);
                const match = player.substring(queryIndex, queryIndex + query.length);
                const after = player.substring(queryIndex + query.length);
                
                item.innerHTML = `${before}<span class="autocomplete-highlight">${match}</span>${after}`;
            } else {
                item.textContent = player;
            }
            
            // Add hover effects and click handlers
            item.addEventListener('mouseenter', () => {
                this.clearSelectedItem();
                this.selectedIndex = index;
                item.classList.add('selected');
            });
            
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
     * Clear selected autocomplete item
     */
    clearSelectedItem() {
        const items = this.autocompleteList.querySelectorAll('.autocomplete-item');
        items.forEach(item => item.classList.remove('selected'));
    }

    /**
     * Navigate autocomplete with arrow keys
     */
    navigateAutocomplete(direction) {
        const items = this.autocompleteList.querySelectorAll('.autocomplete-item');
        if (items.length === 0) return;

        this.clearSelectedItem();

        this.selectedIndex += direction;
        
        if (this.selectedIndex < 0) {
            this.selectedIndex = items.length - 1;
        } else if (this.selectedIndex >= items.length) {
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
            // Automatically submit the guess after selection
            this.makeGuess();
        }
    }

    /**
     * Make a guess with enhanced error handling
     */
    async makeGuess() {
        if (!this.isGameActive || this.isTransitioning) return;

        const playerName = this.guessInput.value.trim();
        if (!playerName) {
            this.showUserFriendlyError('Veuillez entrer le nom d\'un joueur');
            this.guessInput.focus();
            return;
        }

        // Validate session exists
        if (!this.sessionId) {
            this.showUserFriendlyError('Session invalide. Veuillez redémarrer le jeu.');
            return;
        }

        // Show loading state with visual feedback
        this.setLoadingState(true);

        try {
            
            const response = await fetch('/api/guess', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    sessionId: this.sessionId,
                    playerName: playerName
                }),
                timeout: 10000 // 10 second timeout
            });


            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`Erreur serveur: ${response.status} - ${errorText}`);
            }

            const data = await response.json();
            
            if (data.success) {
                this.handleGuessResult(data);
            } else {
                this.handleApiError(data.message || 'Erreur inconnue');
            }
        } catch (error) {
            console.error('Error making guess:', error);
            this.handleNetworkError(error);
        }

        // Reset button state
        this.setLoadingState(false);
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

        // Check if correct - Handle correct guess flow
        if (result.correct) {
            this.handleCorrectGuess(result);
        }

        // Check if game is over (timer ran out or all players found)
        if (result.gameOver) {
            setTimeout(() => {
                this.handleGameOver();
            }, result.correct ? 3000 : 1000);
        }

        this.guessCount++;
        
    }

    /**
     * Handle correct guess flow - Continue with same timer
     */
    handleCorrectGuess(result) {
        
        this.playersFound++;
        
        // Set transitioning state and disable input
        this.isTransitioning = true;
        this.setInputDisabled(true);
        
        // Show success message "Bravo!" for 1 second
        this.showSuccessMessage();
        
        // Move to next player after 1.5 seconds, keeping timer running
        setTimeout(() => {
            this.fadeOutGameState();
            setTimeout(() => {
                this.moveToNextPlayer();
                // Re-enable input after transition
                this.isTransitioning = false;
                this.setInputDisabled(false);
            }, 300);
        }, 1500);
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
        } else {
            console.error('Success overlay element not found!');
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
     * Get square CSS class for coloring
     */
    getSquareClass(comparisonResult) {
        switch (comparisonResult) {
            case 'exact':
                return 'correct';
            case 'partial':
                return 'partial';
            case 'higher':
            case 'lower':
            case 'wrong':
            default:
                return 'wrong';
        }
    }

    /**
     * Create a guess square for the Wordle-style grid
     */
    createGuessSquare(attributeKey, value, comparison, player) {
        const square = document.createElement('div');
        square.className = 'guess-square';
        square.style.transform = 'rotateY(0deg)';
        square.style.transition = 'transform 0.3s ease';
        
        const content = document.createElement('div');
        content.className = 'square-content';
        
        // Handle different attribute types
        switch (attributeKey) {
            case 'name':
                content.innerHTML = `
                    <div class="square-text">${value}</div>
                `;
                break;
                
            case 'team':
                const teamImg = `/assets/teams/${value.replace(/\s+/g, '_')}.png`;
                content.innerHTML = `
                    <img src="${teamImg}" alt="${value}" class="square-image" onerror="this.style.display='none'">
                    <div class="square-text">${this.truncateText(value, 12)}</div>
                `;
                break;
                
            case 'age':
                content.innerHTML = `
                    <div class="square-text">${value}</div>
                `;
                // Add arrow for higher/lower
                if (comparison === 'higher' || comparison === 'lower') {
                    const arrow = document.createElement('div');
                    arrow.className = 'arrow-indicator';
                    arrow.textContent = comparison === 'higher' ? '↓' : '↑';
                    square.appendChild(arrow);
                }
                break;
                
            case 'role':
                content.innerHTML = `
                    <div class="square-text">${value}</div>
                `;
                break;
                
            case 'country':
                const flag = this.countryFlags[value] || '🌍';
                content.innerHTML = `
                    <div class="country-flag">${flag}</div>
                    <div class="square-secondary">${this.smartTruncateCountry(value)}</div>
                `;
                break;
                
            case 'kda':
                content.innerHTML = `
                    <div class="square-text">${value}</div>
                `;
                // Add arrow for higher/lower
                if (comparison === 'higher' || comparison === 'lower') {
                    const arrow = document.createElement('div');
                    arrow.className = 'arrow-indicator';
                    arrow.textContent = comparison === 'higher' ? '↓' : '↑';
                    square.appendChild(arrow);
                }
                break;
                
            case 'champion':
                const champImg = this.getChampionImageUrl(value);
                content.innerHTML = `
                    <img src="${champImg}" alt="${value}" class="square-image" onerror="this.style.display='none'">
                    <div class="square-secondary">${this.truncateText(value, 10)}</div>
                `;
                break;
                
            default:
                content.innerHTML = `
                    <div class="square-text">${this.truncateText(value, 8)}</div>
                `;
        }
        
        square.appendChild(content);
        return square;
    }

    /**
     * Truncate text for square display
     */
    truncateText(text, maxLength) {
        if (text.length <= maxLength) return text;
        return text.substring(0, maxLength - 1) + '…';
    }

    /**
     * Smart truncation for country names
     */
    smartTruncateCountry(country) {
        // Country abbreviations for common long names
        const countryAbbrev = {
            'Czech Republic': 'Czechia',
            'United States': 'USA',
            'United Kingdom': 'UK',
            'South Korea': 'Korea'
        };

        // Use abbreviation if available
        if (countryAbbrev[country]) {
            return countryAbbrev[country];
        }

        // For other countries, use longer limit since columns are now equal size
        return this.truncateText(country, 12);
    }

    /**
     * Add guess result as Wordle-style row with colored squares
     */
    addGuessToHistory(result) {
        const guessRow = document.createElement('div');
        guessRow.className = 'guess-row';
        
        const player = result.comparison.guessed_player;
        const comparisons = result.comparison.comparisons;
        
        // Create squares for each attribute in the correct order
        const attributes = [
            { 
                key: 'name', 
                value: player.player_username,
                comparison: result.correct ? 'exact' : 'wrong'
            },
            { 
                key: 'team', 
                value: player.player_team,
                comparison: comparisons.team || 'wrong'
            },
            { 
                key: 'age', 
                value: player.player_age.toString(),
                comparison: comparisons.age || 'wrong'
            },
            { 
                key: 'role', 
                value: player.player_role,
                comparison: comparisons.role || 'wrong'
            },
            { 
                key: 'country', 
                value: player.player_country,
                comparison: comparisons.country || 'wrong'
            },
            { 
                key: 'kda', 
                value: player.kda_ratio.toFixed(2),
                comparison: comparisons.kda || 'wrong'
            },
            { 
                key: 'champion', 
                value: player.player_most_played_champion,
                comparison: comparisons.champion || 'wrong'
            }
        ];
        
        attributes.forEach(attr => {
            const square = this.createGuessSquare(attr.key, attr.value, attr.comparison, player);
            guessRow.appendChild(square);
        });
        
        this.guessRowsElement.appendChild(guessRow);
        
        // Animate squares one by one
        const squares = guessRow.querySelectorAll('.guess-square');
        squares.forEach((square, index) => {
            setTimeout(() => {
                square.style.transform = 'rotateY(180deg)';
                setTimeout(() => {
                    square.classList.add(this.getSquareClass(attributes[index].comparison));
                    square.style.transform = 'rotateY(0deg)';
                }, 150);
            }, index * 100);
        });
    }


    /**
     * Move to next player - Load next player and clear guess grid
     */
    moveToNextPlayer() {
        
        this.currentPlayer++;
        this.guessCount = 0;
        
        
        // Clear guess grid for next player
        this.guessRowsElement.innerHTML = '';

        // Update player counter
        this.updatePlayerCounter();

        // Check if we've reached the end of players
        if (this.currentPlayer > this.totalPlayers) {
            this.handleGameOver();
            return;
        }

    }

    /**
     * Handle time up - Game over when 2 minutes are up
     */
    handleTimeUp() {
        console.log('Game time is up! 2 minutes elapsed.');
        this.handleGameOver();
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
    async handleGameOver() {
        this.isGameActive = false;
        window.timerManager.stop();
        
        
        // Notify backend that game is over
        try {
            const response = await fetch('/api/end-game', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    sessionId: this.sessionId
                })
            });

            const data = await response.json();
            if (data.success) {
            } else {
                console.error('Failed to mark game as completed:', data.message);
            }
        } catch (error) {
            console.error('Error calling end-game API:', error);
        }
        
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
     * Task 29: Build restart functionality - Simply redirect to home page
     */
    restartGame() {
        
        // Clear session storage
        sessionStorage.removeItem('sessionId');
        
        // Redirect to home page for a fresh start
        window.location.href = '/';
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