// Timer system for Prodle game - 2 minute countdown per player
class TimerManager {
    constructor() {
        this.timerElement = document.getElementById('timer');
        this.timeLeft = 120; // 2 minutes in seconds
        this.intervalId = null;
        this.isRunning = false;
        this.onTimeUp = null;
        this.onTick = null;
        
        // Initialize display
        this.updateDisplay();
    }

    /**
     * Start the 2-minute countdown timer
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
        
        console.log('Starting 2-minute timer...');
        
        // Start the interval
        this.intervalId = setInterval(() => {
            this.tick();
        }, 1000);
        
        // Update display immediately
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
     * Reset timer to 2 minutes for next player
     */
    reset() {
        this.stop();
        this.timeLeft = 120;
        this.updateDisplay();
        this.removeWarningClass();
        console.log('Timer reset for next player');
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
        
        // Add warning class when under 30 seconds
        if (this.timeLeft <= 30) {
            this.addWarningClass();
        }
        
        // Call tick callback if provided
        if (this.onTick && typeof this.onTick === 'function') {
            this.onTick(this.timeLeft);
        }
        
        // Check if time is up
        if (this.timeLeft <= 0) {
            this.handleTimeUp();
        }
    }

    /**
     * Handle when timer reaches 0
     */
    handleTimeUp() {
        this.stop();
        this.addWarningClass();
        
        console.log('Time is up!');
        
        // Call time up callback if provided
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
        
        // Format as M:SS
        const formattedTime = `${minutes}:${seconds.toString().padStart(2, '0')}`;
        this.timerElement.textContent = formattedTime;
        
        // Update color based on time remaining
        if (this.timeLeft <= 10) {
            this.timerElement.style.color = '#FF0000';
        } else if (this.timeLeft <= 30) {
            this.timerElement.style.color = '#FF4444';
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
     * Remove warning class
     */
    removeWarningClass() {
        if (this.timerElement) {
            this.timerElement.classList.remove('warning');
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
     * Get formatted time left as string
     * @returns {string} Formatted time (M:SS)
     */
    getFormattedTimeLeft() {
        const minutes = Math.floor(this.timeLeft / 60);
        const seconds = this.timeLeft % 60;
        return `${minutes}:${seconds.toString().padStart(2, '0')}`;
    }

    /**
     * Check if timer is currently running
     * @returns {boolean} True if timer is running
     */
    isTimerRunning() {
        return this.isRunning;
    }

    /**
     * Set a custom time (useful for testing or special cases)
     * @param {number} seconds - Time in seconds
     */
    setTime(seconds) {
        this.timeLeft = Math.max(0, seconds);
        this.updateDisplay();
    }

    /**
     * Add time to current timer (for bonuses, etc.)
     * @param {number} seconds - Seconds to add
     */
    addTime(seconds) {
        this.timeLeft = Math.max(0, this.timeLeft + seconds);
        this.updateDisplay();
        console.log(`Added ${seconds} seconds to timer`);
    }
}

// Global timer manager instance
window.timerManager = new TimerManager();

// Handle page visibility changes to pause/resume timer
document.addEventListener('visibilitychange', function() {
    if (document.hidden) {
        // Page is hidden, pause timer
        if (window.timerManager && window.timerManager.isTimerRunning()) {
            window.timerManager.pause();
            console.log('Timer paused due to page being hidden');
        }
    } else {
        // Page is visible again, resume timer
        if (window.timerManager && !window.timerManager.isTimerRunning() && window.timerManager.getTimeLeft() > 0) {
            window.timerManager.resume();
            console.log('Timer resumed due to page being visible');
        }
    }
});

// Prevent accidental page refresh during game
window.addEventListener('beforeunload', function(e) {
    if (window.timerManager && window.timerManager.isTimerRunning()) {
        e.preventDefault();
        e.returnValue = 'Une partie est en cours. Êtes-vous sûr de vouloir quitter ?';
        return e.returnValue;
    }
});

// Initialize timer when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    console.log('Timer system initialized');
});