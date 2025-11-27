const form = document.getElementById('queryForm');
const submitBtn = document.getElementById('submitBtn');
const responseContent = document.getElementById('responseContent');
const responseContainer = document.getElementById('responseContainer');
const followUpsContainer = document.getElementById('followUpsContainer');
const errorContainer = document.getElementById('errorContainer');
const statusIndicator = document.getElementById('statusIndicator');
const statusText = document.getElementById('statusText');
const streamingStatusContainer = document.getElementById('streamingStatusContainer');

let currentWebSocket = null;

// Check API status
async function checkApiStatus() {
    try {
        const response = await fetch('/status');
        if (response.ok) {
            statusIndicator.classList.add('online');
            statusText.textContent = 'API Online';
        } else {
            throw new Error('API not responding');
        }
    } catch (error) {
        statusIndicator.classList.remove('online');
        statusText.textContent = 'API Offline';
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    checkApiStatus();
    
    // Configure marked.js if available
    if (typeof marked !== 'undefined') {
        marked.setOptions({
            breaks: true,
            gfm: true
        });
    }
});

// Add streaming status update function
function addStreamingStatus(message, type = 'info') {
    const statusItem = document.createElement('div');
    statusItem.className = `streaming-status-item ${type}`;
    statusItem.innerHTML = `
        <span class="status-icon">${getStatusIcon(type)}</span>
        <span class="status-text">${message}</span>
    `;
    streamingStatusContainer.appendChild(statusItem);
    streamingStatusContainer.scrollTop = streamingStatusContainer.scrollHeight;
}

function getStatusIcon(type) {
    switch(type) {
        case 'success': return '✓';
        case 'error': return '✗';
        case 'warning': return '⚠';
        case 'processing': return '⟳';
        default: return '•';
    }
}

function clearStreamingStatus() {
    streamingStatusContainer.innerHTML = '';
}

// WebSocket connection handler
function connectWebSocket(query, platform, model, previousConversation) {
    // Close existing connection if any
    if (currentWebSocket) {
        currentWebSocket.close();
        currentWebSocket = null;
    }

    // Determine protocol (ws or wss based on current page protocol)
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/query`;

    currentWebSocket = new WebSocket(wsUrl);

    // Set up timeout for connection
    const connectionTimeout = setTimeout(() => {
        if (currentWebSocket && currentWebSocket.readyState !== WebSocket.OPEN) {
            currentWebSocket.close();
            addStreamingStatus('Connection timeout', 'error');
            showError('Failed to establish WebSocket connection');
        }
    }, 10000); // 10 second timeout

    currentWebSocket.onopen = function() {
        clearTimeout(connectionTimeout);
        addStreamingStatus('Connected to server', 'success');
        
        // Send query
        const payload = {
            query: query,
            platform: platform,
            model: model,
            previousConversation: previousConversation || 'No previous conversation provided'
        };
        
        try {
            currentWebSocket.send(JSON.stringify(payload));
        } catch (error) {
            console.error('Error sending message:', error);
            addStreamingStatus('Failed to send query', 'error');
        }
    };

    currentWebSocket.onmessage = function(event) {
        try {
            const message = JSON.parse(event.data);
            handleStreamMessage(message);
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
            addStreamingStatus('Invalid message received', 'warning');
        }
    };

    currentWebSocket.onerror = function(error) {
        console.error('WebSocket error:', error);
        addStreamingStatus('Connection error occurred', 'error');
        showError('WebSocket connection error');
    };

    currentWebSocket.onclose = function(event) {
        clearTimeout(connectionTimeout);
        
        // Log close reason
        if (event.code === 1000) {
            addStreamingStatus('Connection closed normally', 'info');
        } else if (event.code === 1006) {
            addStreamingStatus('Connection closed abnormally', 'warning');
        } else {
            addStreamingStatus(`Connection closed (code: ${event.code})`, 'info');
        }
        
        submitBtn.disabled = false;
        submitBtn.textContent = 'Submit Query';
        currentWebSocket = null;
    };
}

// Cleanup function to close WebSocket when page is unloaded
window.addEventListener('beforeunload', function() {
    if (currentWebSocket) {
        currentWebSocket.close(1000, 'Page unload');
        currentWebSocket = null;
    }
});

// Handle different types of streaming messages
function handleStreamMessage(message) {
    console.log('Received message:', message);
    
    switch(message.type) {
        case 'started':
            addStreamingStatus(message.message, 'success');
            break;
            
        case 'status':
            addStreamingStatus(message.message, 'processing');
            break;
            
        case 'intents':
            const intentNames = message.data.map(intent => intent.Name).join(', ');
            addStreamingStatus(`Intents: ${intentNames}`, 'success');
            break;
            
        case 'vectors':
            if (message.data && message.data.length > 0) {
                addStreamingStatus(`Found ${message.data.length} relevant vectors: [${message.data.join(', ')}]`, 'success');
            } else {
                addStreamingStatus('No additional vectors needed', 'info');
            }
            break;
            
        case 'warning':
            addStreamingStatus(message.message, 'warning');
            break;
            
        case 'response':
            addStreamingStatus('Response received', 'success');
            displayResponse(message.data);
            break;
            
        case 'complete':
            addStreamingStatus(message.message, 'success');
            responseContainer.classList.remove('loading');
            break;
            
        case 'error':
            addStreamingStatus(`Error: ${message.message}`, 'error');
            showError(message.message);
            responseContainer.classList.remove('loading');
            break;
            
        default:
            console.log('Unknown message type:', message.type);
    }
}

form.addEventListener('submit', async (e) => {
    e.preventDefault();

    const query = document.getElementById('query').value.trim();
    const platform = document.getElementById('platform').value;
    const apiVersion = document.getElementById('apiVersion').value;
    const model = document.getElementById('model').value;
    const previousConversation = document.getElementById('previousConversation').value.trim();

    if (!query) {
        showError('Please enter a query');
        return;
    }

    // Show loading state
    submitBtn.disabled = true;
    submitBtn.textContent = 'Processing...';
    responseContainer.className = 'response-area loading';
    responseContent.innerHTML = 'Processing request<span class="spinner"></span>';
    clearMessages();
    clearStreamingStatus();

    // Check if streaming is enabled (v2 with WebSocket)
    const useStreaming = document.getElementById('useStreaming').checked;
    
    if (useStreaming && apiVersion === 'v2') {
        // Use WebSocket for streaming
        addStreamingStatus('Initializing streaming connection...', 'info');
        connectWebSocket(query, platform, model, previousConversation);
    } else {
        // Use traditional HTTP request
        try {
            const requestBody = {
                query: query,
                model: model
            };

            if (previousConversation) {
                try {
                    requestBody.previousConversation = JSON.parse(previousConversation);
                } catch (parseError) {
                    requestBody.previousConversation = previousConversation;
                }
            }

            const response = await fetch(`/handle-query-${apiVersion}?platform=${encodeURIComponent(platform)}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(requestBody)
            });

            const data = await response.json();

            if (response.ok) {
                displayResponse(data.content.result || data.content);
            } else {
                showError(data.message || 'An error occurred while processing your query');
            }
        } catch (error) {
            showError('Network error: ' + error.message);
        } finally {
            // Reset button state
            submitBtn.disabled = false;
            submitBtn.textContent = 'Submit Query';
        }
    }
});

function displayResponse(content) {
    responseContainer.className = 'response-area';
    
    try {
        // Check if content is an object with result property
        let resultContent = content;
        if (typeof content === 'object' && content !== null) {
            resultContent = content.result || JSON.stringify(content, null, 2);
        }
        
        // Render markdown as HTML
        responseContent.innerHTML = renderMarkdown(resultContent);

        // Display follow-ups if available
        if (content.followUps && content.followUps.length > 0) {
            displayFollowUps(content.followUps);
        }
    } catch (error) {
        console.error('Error displaying response:', error);
        // If not JSON, display as markdown
        responseContent.innerHTML = renderMarkdown(String(content));
    }
}

function renderMarkdown(text) {
    // Check if marked library is loaded
    if (typeof marked !== 'undefined' && marked.parse) {
        try {
            return marked.parse(text);
        } catch (e) {
            console.error('Markdown parsing error:', e);
            return escapeHtml(text);
        }
    }
    // Fallback: escape HTML and preserve line breaks
    return escapeHtml(text).replace(/\n/g, '<br>');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function displayFollowUps(followUps) {
    followUpsContainer.innerHTML = `
        <div class="follow-ups">
            <div class="follow-ups-title">Suggested follow-up questions</div>
            ${followUps.map((followUp, index) => `
                <div class="follow-up-item" data-question="${followUp}">
                    ${followUp}
                </div>
            `).join('')}
        </div>
    `;

    // Add event listener for follow-up items
    followUpsContainer.addEventListener('click', function(e) {
        if (e.target.classList.contains('follow-up-item')) {
            const question = e.target.getAttribute('data-question');
            useFollowUp(question);
        }
    });
}

function useFollowUp(question) {
    document.getElementById('query').value = question;
    document.getElementById('query').focus();
}

function showError(message) {
    responseContainer.className = 'response-area empty';
    responseContent.textContent = 'Response will appear here...';
    
    errorContainer.innerHTML = `
        <div class="error">
            ${message}
        </div>
    `;
}

function clearMessages() {
    followUpsContainer.innerHTML = '';
    errorContainer.innerHTML = '';
}
