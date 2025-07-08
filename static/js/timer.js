
class TimerManager {
    constructor() {
        this.timerElement = document.getElementById('timer');
        this.totalGameTime = 120; 
        this.timeLeft = this.totalGameTime;
        this.intervalId = null;
        this.isRunning = false;
        this.onTimeUp = null;
        this.onTick = null;
        
        
        this.loadGameConfig();
        
        
        this.updateDisplay();
    }

    /**
     * Load game configuration from backend
     */
    async loadGameConfig() {
        try {
            const response = await fetch('/api/config');
            const data = await response.json();
            
            if (data.success && data.config) {
                this.totalGameTime = data.config.totalGameTimeSeconds;
                this.timeLeft = this.totalGameTime;
                console.log(`DEBUG: Loaded game config - Total time: ${this.totalGameTime} seconds`);
                this.updateDisplay();
            }
        } catch (error) {
            console.error('Failed to load game config, using default 120 seconds:', error);
        }
    }

    /**
     * Start the total game countdown timer
     * @param {Function} onTimeUp - Callback when timer reaches 0
     * @param {Function} onTick - Callback on each second tick
     */
    start(onTimeUp = null, onTick = null) {
        if (this.isRunning) {
            console.warn('Timer is already running');
            return;
        }

        this.onTimeUp = onTimeUp;
        this.onTick = onTick;
        this.isRunning = true;
        
        console.log(`DEBUG: Starting total game timer - ${this.totalGameTime} seconds...`);
        
        
        this.intervalId = setInterval(() => {
            this.tick();
        }, 1000);
        
        
        this.updateDisplay();
    }

    /**
     * Stop the timer
     */
    stop() {
        if (this.intervalId) {
            clearInterval(this.intervalId);
            this.intervalId = null;
        }
        this.isRunning = false;
        console.log('Timer stopped');
    }

    /**
     * Reset timer to full game time (for new game)
     */
    reset() {
        this.stop();
        this.timeLeft = this.totalGameTime;
        this.updateDisplay();
        this.removeWarningClass();
        console.log(`DEBUG: Timer reset to ${this.totalGameTime} seconds for new game`);
    }

    /**
     * Pause the timer
     */
    pause() {
        if (this.isRunning) {
            this.stop();
            console.log('Timer paused');
        }
    }

    /**
     * Resume the timer
     */
    resume() {
        if (!this.isRunning && this.timeLeft > 0) {
            this.start(this.onTimeUp, this.onTick);
            console.log('Timer resumed');
        }
    }

    /**
     * Handle each second tick
     */
    tick() {
        this.timeLeft--;
        this.updateDisplay();
        
        
        if (this.timeLeft <= 30) {
            this.addWarningClass();
        }
        
        
        if (this.onTick && typeof this.onTick === 'function') {
            this.onTick(this.timeLeft);
        }
        
        
        if (this.timeLeft <= 0) {
            this.handleTimeUp();
        }
    }

    /**
     * Handle when timer reaches 0
     */
    handleTimeUp() {
        this.stop();
        this.addCriticalClass();
        
        console.log('Time is up!');
        
        
        if (this.onTimeUp && typeof this.onTimeUp === 'function') {
            this.onTimeUp();
        }
    }

    /**
     * Update the timer display
     */
    updateDisplay() {
        if (!this.timerElement) {
            console.error('Timer element not found');
            return;
        }

        const minutes = Math.floor(this.timeLeft / 60);
        const seconds = this.timeLeft % 60;
        
        
        const formattedTime = `${minutes}:${seconds.toString().padStart(2, '0')}`;
        this.timerElement.textContent = formattedTime;
        
        
        this.timerElement.classList.remove('warning', 'critical');
        
        if (this.timeLeft <= 10) {
            this.timerElement.style.color = '#FF0000';
            this.timerElement.classList.add('critical');
        } else if (this.timeLeft <= 30) {
            this.timerElement.style.color = '#FF4444';
            this.timerElement.classList.add('warning');
        } else {
            this.timerElement.style.color = '#FF6B35';
        }
    }

    /**
     * Add warning class for visual indication
     */
    addWarningClass() {
        if (this.timerElement) {
            this.timerElement.classList.add('warning');
        }
    }

    /**
     * Add critical class for urgent visual indication
     */
    addCriticalClass() {
        if (this.timerElement) {
            this.timerElement.classList.add('critical');
        }
    }

    /**
     * Remove warning class
     */
    removeWarningClass() {
        if (this.timerElement) {
            this.timerElement.classList.remove('warning', 'critical');
        }
    }

    /**
     * Get current time left in seconds
     * @returns {number} Time left in seconds
     */
    getTimeLeft() {
        return this.timeLeft;
    }


    /**
     * Check if timer is currently running
     * @returns {boolean} True if timer is running
     */
    isTimerRunning() {
        return this.isRunning;
    }

}


window.timerManager = new TimerManager();



window.addEventListener('beforeunload', function(e) {
    if (window.timerManager && window.timerManager.isTimerRunning()) {
        e.preventDefault();
        e.returnValue = 'Une partie est en cours. Êtes-vous sûr de vouloir quitter ?';
        return e.returnValue;
    }
});

