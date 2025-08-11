// Global functions for device interactions
function openDeviceModal(deviceId) {
    fetch(`/ui/device/${deviceId}`)
        .then(response => response.text())
        .then(html => {
            const modal = document.getElementById('device-modal');
            modal.innerHTML = html;
            modal.classList.remove('hidden');

            // Initialize Alpine component
            Alpine.initTree(modal);
        })
        .catch(error => {
            console.error('Error loading device details:', error);
            showNotification('Failed to load device details', 'error');
        });
}

function simulateSale(deviceId) {
    fetch(`/server/v1/simulation/device/${deviceId}/sale`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            quantity: Math.floor(Math.random() * 5) + 1
        })
    })
        .then(response => {
            if (response.ok) {
                showNotification('Sale simulated successfully', 'success');
            } else {
                showNotification('Failed to simulate sale', 'error');
            }
        })
        .catch(error => {
            console.error('Error simulating sale:', error);
            showNotification('Failed to simulate sale', 'error');
        });
}

function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    const bgColor = type === 'success' ? 'bg-green-500' : type === 'error' ? 'bg-red-500' : 'bg-blue-500';

    notification.className = `fixed top-4 right-4 ${bgColor} text-white px-6 py-3 rounded-lg shadow-lg z-50 transform transition-all duration-300 translate-x-full`;
    notification.textContent = message;

    document.body.appendChild(notification);

    // Slide in
    setTimeout(() => {
        notification.classList.remove('translate-x-full');
    }, 100);

    // Slide out and remove
    setTimeout(() => {
        notification.classList.add('translate-x-full');
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    }, 3000);
}

// HTMX configuration
document.body.addEventListener('htmx:afterSwap', function(evt) {
    // Re-initialize any Alpine components in swapped content
    if (typeof Alpine !== 'undefined') {
        Alpine.initTree(evt.detail.target);
    }
});

// Handle HTMX redirects explicitly
document.body.addEventListener('htmx:responseError', function(evt) {
    console.log('HTMX response error:', evt.detail);
});

document.body.addEventListener('htmx:beforeRedirect', function(evt) {
    console.log('HTMX redirect to:', evt.detail.path);
});

document.body.addEventListener('htmx:afterRequest', function(evt) {
    // Check for HX-Redirect header and handle it manually if needed
    const redirectTo = evt.detail.xhr.getResponseHeader('HX-Redirect');
    if (redirectTo) {
        console.log('Manual redirect to:', redirectTo);
        window.location.href = redirectTo;
    }
});

// WebSocket reconnection logic
let reconnectInterval;
function setupWebSocketReconnect() {
    if (window.ws && window.ws.readyState === WebSocket.CLOSED) {
        clearInterval(reconnectInterval);
        reconnectInterval = setInterval(() => {
            if (window.ws.readyState === WebSocket.CLOSED) {
                window.location.reload();
            }
        }, 5000);
    }
}

// Check WebSocket connection every 10 seconds
setInterval(setupWebSocketReconnect, 10000);