class GameManager {
    constructor() {
        this.sessionId = '';
        this.currentPlayer = 1;
        this.totalPlayers = 20;
        this.score = 0;
        this.guessCount = 0;
        this.isGameActive = false;
        this.currentTargetPlayer = null;
        this.playersFound = 0;
        this.isTransitioning = false;
        
        this.guessInput = document.getElementById('guess-input');
        this.guessButton = document.getElementById('guess-button');
        this.scoreElement = document.getElementById('score');
        this.playerCounterElement = document.getElementById('player-counter');
        this.guessRowsElement = document.getElementById('guess-rows');
        this.autocompleteList = document.getElementById('autocomplete-list');
        
        this.successOverlay = document.getElementById('success-overlay');
        this.endGameOverlay = document.getElementById('end-game-overlay');
        this.loadingOverlay = document.getElementById('loading-overlay');
        this.finalScoreElement = document.getElementById('final-score');
        this.playersCompletedElement = document.getElementById('players-completed');
        this.usernameInput = document.getElementById('username-input');
        this.scoreForm = document.getElementById('score-form');
        this.scoreSubmitted = document.getElementById('score-submitted');
        this.submitScoreBtn = document.getElementById('submit-score-btn');
        this.playerRankElement = document.getElementById('player-rank');
        
        this.selectedIndex = -1;
        this.autocompleteResults = [];
        this.autocompleteDebounceTimer = null;
        this.debounceDelay = 10;
        
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

    initialize() {
        this.sessionId = sessionStorage.getItem('sessionId') || document.getElementById('session-id').value;
        
        if (!this.sessionId) {
            console.error('No session ID found! Redirecting to home page.');
            window.location.href = '/';
            return;
        }
        
        document.getElementById('session-id').value = this.sessionId;
        
        this.isGameActive = true;
        
        window.timerManager.onTimeUp = () => this.handleTimeUp();
        window.timerManager.onTick = (timeLeft) => this.handleTimerTick(timeLeft);
        
        this.loadLeaderboard();
        
        console.log('Game initialized');
    }

    setupInitialState() {
        this.sessionId = sessionStorage.getItem('sessionId') || document.getElementById('session-id').value;
        
        if (!this.sessionId) {
            console.error('No session ID found during setup! Redirecting to home page.');
            window.location.href = '/';
            return;
        }
        
        document.getElementById('session-id').value = this.sessionId;
        
        this.isGameActive = false;
        
        if (this.guessInput) this.guessInput.disabled = true;
        if (this.guessButton) this.guessButton.disabled = true;
        
        console.log('Game state initialized with session ID:', this.sessionId);
    }

    setupEventListeners() {
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

        if (this.guessButton) {
            this.guessButton.addEventListener('click', () => this.makeGuess());
        }

        if (this.usernameInput) {
            this.usernameInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    this.submitFinalScore();
                }
            });
        }
    }

    showLoading(text = 'Chargement...') {
        if (this.loadingOverlay) {
            this.loadingOverlay.querySelector('.loading-text').textContent = text;
            this.loadingOverlay.classList.remove('hidden');
        }
    }

    hideLoading() {
        if (this.loadingOverlay) {
            this.loadingOverlay.classList.add('hidden');
        }
    }

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
        
        if (disabled) {
            this.hideAutocomplete();
        }
    }

    showUserFriendlyError(message) {
        
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

        setTimeout(() => {
            errorEl.style.animation = 'slideUp 0.3s ease forwards';
            setTimeout(() => errorEl.remove(), 300);
        }, 4000);
    }

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

    handleNetworkError(error) {
        if (error.name === 'TypeError' && error.message.includes('fetch')) {
            this.showUserFriendlyError('Problème de connexion. Vérifiez votre connexion internet.');
        } else if (error.message.includes('timeout')) {
            this.showUserFriendlyError('Délai d\'attente dépassé. Veuillez réessayer.');
        } else {
            this.showUserFriendlyError('Erreur de connexion au serveur');
        }
    }

    handleInputChange(event) {
        if (this.isTransitioning || !this.isGameActive) return;
        
        const query = event.target.value.trim();
        
        if (this.autocompleteDebounceTimer) {
            clearTimeout(this.autocompleteDebounceTimer);
        }
        
        if (query.length >= 2) {
            this.guessInput.classList.add('loading');
            
            this.autocompleteDebounceTimer = setTimeout(async () => {
                if (!this.sessionId) {
                    console.warn('Session ID lost during autocomplete request');
                    this.guessInput.classList.remove('loading');
                    return;
                }
                
                await this.fetchAutocomplete(query);
                this.guessInput.classList.remove('loading');
            }, this.debounceDelay);
        } else {
            this.hideAutocomplete();
            this.guessInput.classList.remove('loading');
        }
    }

    handleKeyDown(event) {
        if (event.key === 'Enter') {
            event.preventDefault();
            if (this.selectedIndex >= 0 && this.autocompleteResults.length > 0) {
                this.selectAutocompleteItem(this.selectedIndex);
            } else if (this.autocompleteResults.length > 0) {
                this.selectAutocompleteItem(0);
            } else {
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

    async fetchAutocomplete(query) {
        try {
            if (!this.sessionId) {
                console.warn('No session ID available for autocomplete request');
                this.hideAutocomplete();
                return;
            }
            
            const url = `/api/autocomplete?query=${encodeURIComponent(query)}&sessionId=${encodeURIComponent(this.sessionId)}`;
            const response = await fetch(url);
            const data = await response.json();
            
            this.autocompleteResults = data.players || [];
            this.selectedIndex = -1;
            this.showAutocomplete();
        } catch (error) {
            console.error('Error fetching autocomplete:', error);
            this.hideAutocomplete();
        }
    }

    showAutocomplete() {
        const query = this.guessInput.value.trim().toLowerCase();
        
        if (this.autocompleteResults.length === 0) {
            this.hideAutocomplete();
            return;
        }

        this.autocompleteList.innerHTML = '';
        
        const limitedResults = this.autocompleteResults.slice(0, 10);
        
        limitedResults.forEach((player, index) => {
            const item = document.createElement('div');
            item.className = 'autocomplete-item';
            
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

    hideAutocomplete() {
        setTimeout(() => {
            this.autocompleteList.classList.add('hidden');
        }, 150);
    }

    clearSelectedItem() {
        const items = this.autocompleteList.querySelectorAll('.autocomplete-item');
        items.forEach(item => item.classList.remove('selected'));
    }

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

    selectAutocompleteItem(index) {
        if (index >= 0 && index < this.autocompleteResults.length) {
            this.guessInput.value = this.autocompleteResults[index];
            this.hideAutocomplete();
            this.guessInput.focus();
            
            this.makeGuess();
        }
    }

    async makeGuess() {
        if (!this.isGameActive || this.isTransitioning) return;

        const playerName = this.guessInput.value.trim();
        if (!playerName) {
            this.showUserFriendlyError('Veuillez entrer le nom d\'un joueur');
            this.guessInput.focus();
            return;
        }

        if (!this.sessionId) {
            this.showUserFriendlyError('Session invalide. Veuillez redémarrer le jeu.');
            return;
        }

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
                timeout: 10000
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

        this.setLoadingState(false);
    }

    handleGuessResult(result) {
        

        this.score = result.score;
        this.updateScoreDisplay();

        this.addGuessToHistory(result);

        this.guessInput.value = '';
        this.hideAutocomplete();

        if (result.correct) {
            this.handleCorrectGuess(result);
        }

        if (result.gameOver) {
            setTimeout(() => {
                this.handleGameOver();
            }, result.correct ? 3000 : 1000);
        }

        this.guessCount++;
        
    }

    handleCorrectGuess(result) {
        
        this.playersFound++;
        
        this.isTransitioning = true;
        this.setInputDisabled(true);
        
        this.showSuccessMessage();
        
        setTimeout(() => {
            this.fadeOutGameState();
            setTimeout(() => {
                this.moveToNextPlayer();
                this.isTransitioning = false;
                this.setInputDisabled(false);
            }, 300);
        }, 1500);
    }

    showSuccessMessage() {
        if (this.successOverlay) {
            this.successOverlay.classList.remove('hidden');
            
            setTimeout(() => {
                this.successOverlay.classList.add('hidden');
            }, 1000);
        } else {
            console.error('Success overlay element not found!');
        }
    }

    fadeOutGameState() {
        const mainGame = document.querySelector('.main-game');
        if (mainGame) {
            mainGame.style.transition = 'opacity 0.5s ease-out';
            mainGame.style.opacity = '0.3';
            
            setTimeout(() => {
                mainGame.style.opacity = '1';
            }, 1000);
        }
    }

    createPlayerAttributeCard(label, value, comparisonResult, attributeKey, targetValue = null) {
        const card = document.createElement('div');
        card.className = `player-attribute ${comparisonResult}`;
        
        const labelDiv = document.createElement('div');
        labelDiv.className = 'attribute-label';
        labelDiv.textContent = label;
        
        const valueDiv = document.createElement('div');
        valueDiv.className = 'attribute-value';
        
        
        if (attributeKey === 'team') {
            const teamImg = document.createElement('img');
            teamImg.className = 'team-image';
            teamImg.src = `/assets/teams/${value}.png`;
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
        
        if ((attributeKey === 'year_of_birth' || attributeKey === 'last_split_result' || attributeKey === 'first_split_in_league') && comparisonResult === 'incorrect' && targetValue !== null) {
            const arrow = document.createElement('span');
            arrow.className = 'arrow-indicator';
            
            if (attributeKey === 'year_of_birth') {
                const numValue = parseFloat(value);
                const numTarget = parseFloat(targetValue);
                
                if (numValue > numTarget) {
                    arrow.textContent = '↓';
                    arrow.classList.add('arrow-down');
                } else if (numValue < numTarget) {
                    arrow.textContent = '↑';
                    arrow.classList.add('arrow-up');
                }
            } else if (attributeKey === 'last_split_result') {
                const numValue = parseInt(value.replace(/[^\d]/g, ''));
                const numTarget = parseFloat(targetValue);
                
                if (numValue > numTarget) {
                    arrow.textContent = '↑';
                    arrow.classList.add('arrow-up');
                } else if (numValue < numTarget) {
                    arrow.textContent = '↓';
                    arrow.classList.add('arrow-down');
                }
            } else if (attributeKey === 'first_split_in_league') {
                const numValue = parseFloat(value);
                const numTarget = parseFloat(targetValue);
                
                if (numValue > numTarget) {
                    arrow.textContent = '↓';
                    arrow.classList.add('arrow-down');
                } else if (numValue < numTarget) {
                    arrow.textContent = '↑';
                    arrow.classList.add('arrow-up');
                }
            }
            
            card.appendChild(arrow);
        }
        
        card.appendChild(labelDiv);
        card.appendChild(valueDiv);
        
        return card;
    }


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
        
        
        switch (attributeKey) {
            case 'name':
                content.innerHTML = `
                    <div class="square-text">${value}</div>
                `;
                break;
                
            case 'team':
                const teamImg = `/assets/teams/${value}.png`;
                content.innerHTML = `
                    <img src="${teamImg}" alt="${value}" class="square-image" onerror="this.style.display='none'">
                    <div class="square-text">${this.truncateText(value, 12)}</div>
                `;
                break;
                
            case 'year_of_birth':
                content.innerHTML = `
                    <div class="square-text">${value}</div>
                `;
                
                if (comparison === 'higher' || comparison === 'lower') {
                    const arrow = document.createElement('div');
                    arrow.className = 'arrow-indicator';
                    
                    arrow.textContent = comparison === 'higher' ? '↑' : '↓';
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
                
            case 'last_split_result':
                
                content.innerHTML = `
                    <div class="square-text">${value}</div>
                `;
                
                if (comparison === 'higher' || comparison === 'lower') {
                    const arrow = document.createElement('div');
                    arrow.className = 'arrow-indicator';
                    
                    arrow.textContent = comparison === 'higher' ? '↑' : '↓';
                    square.appendChild(arrow);
                }
                break;
                
            case 'first_split_in_league':
                content.innerHTML = `
                    <div class="square-text">${this.truncateText(value, 10)}</div>
                `;
                
                if (comparison === 'higher' || comparison === 'lower') {
                    const arrow = document.createElement('div');
                    arrow.className = 'arrow-indicator';
                    
                    arrow.textContent = comparison === 'higher' ? '↑' : '↓';
                    square.appendChild(arrow);
                }
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
        
        const countryAbbrev = {
            'Czech Republic': 'Czechia',
            'United States': 'USA',
            'United Kingdom': 'UK',
            'South Korea': 'Korea'
        };

        
        if (countryAbbrev[country]) {
            return countryAbbrev[country];
        }

        
        return this.truncateText(country, 12);
    }

    /**
     * Format number as French ordinal (1er, 2e, 3e, etc.)
     */
    formatFrenchOrdinal(num) {
        const number = parseInt(num);
        if (isNaN(number)) return num;
        
        if (number === 1) {
            return '1er';
        } else {
            return number + 'e';
        }
    }

    /**
     * Add guess result as Wordle-style row with colored squares
     */
    addGuessToHistory(result) {
        const guessRow = document.createElement('div');
        guessRow.className = 'guess-row';
        
        const player = result.comparison.guessed_player;
        const comparisons = result.comparison.comparisons;
        
        
        const attributes = [
            { 
                key: 'name', 
                value: player.ID,
                comparison: result.correct ? 'exact' : 'wrong'
            },
            { 
                key: 'team', 
                value: player.Team,
                comparison: comparisons.team || 'wrong'
            },
            { 
                key: 'year_of_birth', 
                value: player.YearOfBirth.toString(),
                comparison: comparisons.year_of_birth || 'wrong'
            },
            { 
                key: 'role', 
                value: player.Role,
                comparison: comparisons.role || 'wrong'
            },
            { 
                key: 'country', 
                value: player.Nationality,
                comparison: comparisons.country || 'wrong'
            },
            { 
                key: 'last_split_result', 
                value: this.formatFrenchOrdinal(player.LastSplitResult),
                comparison: comparisons.last_split_result || 'wrong'
            },
            { 
                key: 'first_split_in_league', 
                value: player.FirstSplitInLeague.toString(),
                comparison: comparisons.first_split_in_league || 'wrong'
            }
        ];
        
        attributes.forEach(attr => {
            const square = this.createGuessSquare(attr.key, attr.value, attr.comparison, player);
            guessRow.appendChild(square);
        });
        
        this.guessRowsElement.appendChild(guessRow);
        
        
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
        
        
        
        this.guessRowsElement.innerHTML = '';

        
        this.updatePlayerCounter();

        
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
        
    }

    /**
     * Task 27: Create end game flow
     */
    async handleGameOver() {
        this.isGameActive = false;
        window.timerManager.stop();
        
        
        
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
        
        
        this.scoreForm.classList.remove('hidden');
        this.scoreSubmitted.classList.add('hidden');
        
        
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
                
                this.showScoreSubmitted(data.rank);
                this.loadLeaderboard(); 
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
     * Show score submitted confirmation with rank
     */
    showScoreSubmitted(rank = null) {
        this.scoreForm.classList.add('hidden');
        this.scoreSubmitted.classList.remove('hidden');
        
        
        if (rank && rank > 0 && this.playerRankElement) {
            let rankText;
            if (rank === 1) {
                rankText = `🏆 Vous êtes #${rank} sur le classement! 🏆`;
            } else if (rank <= 3) {
                rankText = `🥉 Vous êtes #${rank} sur le classement!`;
            } else if (rank <= 10) {
                rankText = `🎉 Vous êtes #${rank} sur le classement!`;
            } else {
                rankText = `Vous êtes #${rank} sur le classement!`;
            }
            
            this.playerRankElement.textContent = rankText;
            this.playerRankElement.classList.remove('hidden');
        }
    }

    /**
     * Task 29: Build restart functionality - Simply redirect to home page
     */
    restartGame() {
        
        
        sessionStorage.removeItem('sessionId');
        
        
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
        
    }
}


window.gameManager = null;


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


document.addEventListener('DOMContentLoaded', function() {
    console.log('Complete game flow system initialized');
    if (!window.gameManager) {
        window.gameManager = new GameManager();
    }
});