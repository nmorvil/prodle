
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

        
        this.countdownNumber.textContent = count;
        
        
        this.countdownNumber.style.animation = 'none';
        this.countdownNumber.offsetHeight; 
        this.countdownNumber.style.animation = 'countdownFade 1s ease-in-out';

        
        if (count === 1) {
            this.countdownNumber.style.color = '#FF4444';
            this.countdownNumber.style.transform = 'scale(1.1)';
        } else {
            this.countdownNumber.style.color = '#FFD700';
            this.countdownNumber.style.transform = 'scale(1)';
        }

        
        setTimeout(() => {
            this.runCountdown(count - 1);
        }, 1000);
    }

}


window.countdownManager = new CountdownManager();


async function createNewSession() {
    console.log('Creating new session...');
    
    try {
        console.log('Current URL:', window.location.href);
        console.log('URL search params:', window.location.search);
        
        
        let difficulty = 'difficile'; 
        try {
            const urlParams = new URLSearchParams(window.location.search);
            difficulty = urlParams.get('difficulty') || 'difficile';
            console.log('Using difficulty from URL:', difficulty);
        } catch (urlError) {
            console.error('Error parsing URL parameters:', urlError);
        }
        
        const requestBody = {
            difficulty: difficulty
        };
        console.log('Sending request body:', JSON.stringify(requestBody));
        console.log('Request body type:', typeof requestBody);
        console.log('Request body stringified length:', JSON.stringify(requestBody).length);
        
        const response = await fetch('/api/start-game', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestBody)
        });

        console.log('Session creation response:', response.status, response.statusText);

        const data = await response.json();
        console.log('Session data received:', data);

        if (!response.ok) {
            console.error('Server returned error:', data);
            throw new Error(`Failed to create new session: ${response.status} ${response.statusText}`);
        }
        
        if (data.success && data.sessionId) {
            console.log('New session created with ID:', data.sessionId);
            
            
            sessionStorage.removeItem('sessionId');
            
            
            sessionStorage.setItem('sessionId', data.sessionId);
            
            
            const sessionInput = document.getElementById('session-id');
            if (sessionInput) {
                sessionInput.value = data.sessionId;
                console.log('Session ID stored in hidden input');
            }

            return data.sessionId;
        } else {
            throw new Error(`Invalid response from server: ${JSON.stringify(data)}`);
        }
    } catch (error) {
        console.error('Error creating new session:', error);
        alert('Erreur lors de la création de la session. Redirection vers l\'accueil...');
        
        window.location.href = '/';
        return null;
    }
}


document.addEventListener('DOMContentLoaded', async function() {
    console.log('DOM Content Loaded - starting session creation...');
    
    
    const sessionId = await createNewSession();
    
    if (sessionId) {
        console.log('Session created successfully, starting game flow...');
        
        
        if (window.gameManager) {
            window.gameManager.setupInitialState();
        }

        
        window.countdownManager.start(() => {
            console.log('Countdown finished, starting 2-minute game timer...');
            
            
            if (window.timerManager) {
                window.timerManager.start(
                    () => {
                        
                        if (window.gameManager) {
                            window.gameManager.handleTimeUp();
                        }
                    },
                    (timeLeft) => {
                        
                        if (window.gameManager) {
                            window.gameManager.handleTimerTick(timeLeft);
                        }
                    }
                );
            }
            
            
            const guessInput = document.getElementById('guess-input');
            const guessButton = document.getElementById('guess-button');
            
            if (guessInput) {
                guessInput.disabled = false;
                guessInput.focus();
            }
            
            if (guessButton) {
                guessButton.disabled = false;
            }

            
            if (window.gameManager) {
                window.gameManager.initialize();
            }
        });
    }
});



window.addEventListener('beforeunload', function(e) {
    if (!window.countdownManager.countdownOverlay.classList.contains('hidden')) {
        e.preventDefault();
        e.returnValue = 'Le compte à rebours est en cours. Êtes-vous sûr de vouloir quitter ?';
        return e.returnValue;
    }
});