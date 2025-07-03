// Countdown functionality for Prodle game
class CountdownManager {
    constructor() {
        this.countdownOverlay = document.getElementById('countdown-overlay');
        this.countdownNumber = document.getElementById('countdown-number');
        this.onComplete = null;
    }

    /**
     * Start the countdown sequence (3, 2, 1)
     * @param {Function} onComplete - Callback function to execute when countdown finishes
     */
    start(onComplete = null) {
        this.onComplete = onComplete;
        this.showCountdown();
        this.runCountdown(3);
    }

    /**
     * Show the countdown overlay
     */
    showCountdown() {
        this.countdownOverlay.classList.remove('hidden');
        this.countdownOverlay.style.display = 'flex';
    }

    /**
     * Hide the countdown overlay
     */
    hideCountdown() {
        this.countdownOverlay.style.display = 'none';
        this.countdownOverlay.classList.add('hidden');
    }

    /**
     * Run the countdown sequence
     * @param {number} count - Current countdown number
     */
    runCountdown(count) {
        if (count <= 0) {
            this.hideCountdown();
            if (this.onComplete && typeof this.onComplete === 'function') {
                this.onComplete();
            }
            return;
        }

        // Update the display
        this.countdownNumber.textContent = count;
        
        // Reset animation by removing and re-adding the class
        this.countdownNumber.style.animation = 'none';
        this.countdownNumber.offsetHeight; // Trigger reflow
        this.countdownNumber.style.animation = 'countdownFade 1s ease-in-out';

        // Add special styling for the last second
        if (count === 1) {
            this.countdownNumber.style.color = '#FF4444';
            this.countdownNumber.style.transform = 'scale(1.1)';
        } else {
            this.countdownNumber.style.color = '#FFD700';
            this.countdownNumber.style.transform = 'scale(1)';
        }

        // Continue countdown after 1 second
        setTimeout(() => {
            this.runCountdown(count - 1);
        }, 1000);
    }

}

// Global countdown manager instance
window.countdownManager = new CountdownManager();

// Auto-start countdown when page loads if we have a session
document.addEventListener('DOMContentLoaded', function() {
    // Check if we have a session ID from the previous page
    const sessionId = sessionStorage.getItem('sessionId');
    
    if (sessionId) {
        // Store session ID in hidden input for other scripts to use
        const sessionInput = document.getElementById('session-id');
        if (sessionInput) {
            sessionInput.value = sessionId;
        }

        // Setup initial game state
        if (window.gameManager) {
            window.gameManager.setupInitialState();
        }

        // Start countdown, then start the game timer
        window.countdownManager.start(() => {
            console.log('Countdown finished, starting 2-minute game timer...');
            
            // Start the single 2-minute timer for entire game
            if (window.timerManager) {
                window.timerManager.start(
                    () => {
                        // Time up callback - end the game
                        if (window.gameManager) {
                            window.gameManager.handleTimeUp();
                        }
                    },
                    (timeLeft) => {
                        // Tick callback - update any UI if needed
                        if (window.gameManager) {
                            window.gameManager.handleTimerTick(timeLeft);
                        }
                    }
                );
            }
            
            // Enable game controls
            const guessInput = document.getElementById('guess-input');
            const guessButton = document.getElementById('guess-button');
            
            if (guessInput) {
                guessInput.disabled = false;
                guessInput.focus();
            }
            
            if (guessButton) {
                guessButton.disabled = false;
            }

            // Initialize game functionality (without disabling controls)
            if (window.gameManager) {
                window.gameManager.initialize();
            }
        });
    } else {
        // No session found, redirect back to home
        console.warn('No session ID found, redirecting to home');
        window.location.href = '/';
    }
});


// Prevent accidental page refresh during countdown
window.addEventListener('beforeunload', function(e) {
    if (!window.countdownManager.countdownOverlay.classList.contains('hidden')) {
        e.preventDefault();
        e.returnValue = 'Le compte à rebours est en cours. Êtes-vous sûr de vouloir quitter ?';
        return e.returnValue;
    }
});