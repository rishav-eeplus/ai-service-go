const form = document.getElementById('queryForm');
const submitBtn = document.getElementById('submitBtn');
const responseContent = document.getElementById('responseContent');
const responseContainer = document.getElementById('responseContainer');
const followUpsContainer = document.getElementById('followUpsContainer');
const errorContainer = document.getElementById('errorContainer');
const statusIndicator = document.getElementById('statusIndicator');
const statusText = document.getElementById('statusText');
const streamingStatusContainer = document.getElementById('streamingStatusContainer');
const inlineStreamingContainer = document.getElementById('inlineStreamingContainer');

let currentEventSource = null;
let currentQuery = '';
let currentPlatform = '';
let currentModel = '';
let currentPreviousConversation = '';
let currentClarificationCount = 0;

// Conversation history tracking
let conversationHistory = [];

// Helper function to get the base path from current URL
function getBasePath() {
    // Remove trailing filename and /ui suffix to get the API base path
    let path = window.location.pathname.replace(/\/[^\/]*$/, ''); // Remove filename
    path = path.replace(/\/ui$/, ''); // Remove /ui suffix
    return path;
}

// Check API status
async function checkApiStatus() {
    try {
        const response = await fetch(`${getBasePath()}/status`);
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
document.addEventListener('DOMContentLoaded', function () {
    checkApiStatus();

    // Configure marked.js if available
    if (typeof marked !== 'undefined') {
        marked.setOptions({
            breaks: true,
            gfm: true
        });
    }
});

// Update the previous conversation textarea with current history
function updatePreviousConversationField() {
    const previousConversationField = document.getElementById('previousConversation');
    if (previousConversationField && conversationHistory.length > 0) {
        previousConversationField.value = JSON.stringify(conversationHistory, null, 2);
    }
}

// Add a message to conversation history
function addToConversationHistory(role, content) {
    conversationHistory.push({ role, content });
    updatePreviousConversationField();
}

// Clear conversation history
function clearConversationHistory() {
    conversationHistory = [];
    const previousConversationField = document.getElementById('previousConversation');
    if (previousConversationField) {
        previousConversationField.value = '';
    }
}

// Add streaming status update function - updates single message instead of stacking
function addStreamingStatus(message, type = 'info') {
    // Update the single status message instead of appending
    inlineStreamingContainer.innerHTML = `
        <div class="streaming-status-item ${type}">
            <span class="status-icon">${getStatusIcon(type)}</span>
            <span class="status-text">${message}</span>
        </div>
    `;
    responseContainer.scrollTop = responseContainer.scrollHeight;
}

function getStatusIcon(type) {
    switch (type) {
        case 'success': return '✓';
        case 'error': return '✗';
        case 'warning': return '⚠';
        case 'processing': return '⟳';
        default: return '•';
    }
}

function clearStreamingStatus() {
    inlineStreamingContainer.innerHTML = '';
}

// SSE connection handler
function connectSSE(query, platform, model, previousConversation, clarificationResponse = '', clarificationCount = 0) {
    // Close existing connection if any
    if (currentEventSource) {
        currentEventSource.close();
        currentEventSource = null;
    }

    // Store current request params for clarification retry
    currentQuery = query;
    currentPlatform = platform;
    currentModel = model;
    currentPreviousConversation = previousConversation;
    currentClarificationCount = clarificationCount;

    addStreamingStatus('Connecting to server...', 'info');

    // Make POST request with fetch
    const requestBody = {
        query: query,
        platform: platform,
        model: model,
        previousConversation: previousConversation || 'No previous conversation provided',
        clarificationResponse: clarificationResponse,
        clarificationCount: clarificationCount
    };

    fetch(`${getBasePath()}/handle-query-v2`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody)
    }).then(response => {
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';

        function processText({ done, value }) {
            if (done) {
                addStreamingStatus('Connection closed', 'info');
                submitBtn.disabled = false;
                submitBtn.textContent = 'Submit Query';
                return;
            }

            buffer += decoder.decode(value, { stream: true });
            const lines = buffer.split('\n');
            buffer = lines.pop(); // Keep the last incomplete line in buffer

            for (const line of lines) {
                if (line.startsWith('data: ')) {
                    try {
                        const data = line.substring(6);
                        const message = JSON.parse(data);
                        handleStreamMessage(message);
                    } catch (error) {
                        console.error('Error parsing SSE message:', error);
                    }
                }
            }

            return reader.read().then(processText);
        }

        return reader.read().then(processText);
    }).catch(error => {
        console.error('SSE connection error:', error);
        addStreamingStatus('Connection error occurred', 'error');
        showError('Failed to connect to server');
        submitBtn.disabled = false;
        submitBtn.textContent = 'Submit Query';
    });
}

// Cleanup function when page is unloaded
window.addEventListener('beforeunload', function () {
    if (currentEventSource) {
        currentEventSource.close();
        currentEventSource = null;
    }
});

// Handle different types of streaming messages
function handleStreamMessage(message) {
    console.log('Received message:', message);

    switch (message.type) {
        case 'started':
            addStreamingStatus(message.message, 'success');
            break;

        case 'info':
            addStreamingStatus(message.message, 'info');
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

        case 'success':
            addStreamingStatus(message.message, 'success');
            break;

        case 'clarification':
            addStreamingStatus('Clarification needed', 'warning');
            // Extract clarificationCount from message data
            let newCount = 0;
            if (message.data && typeof message.data === 'object') {
                try {
                    const clarificationData = typeof message.data === 'string' ? JSON.parse(message.data) : message.data;
                    newCount = clarificationData.clarificationCount || 0;
                } catch (e) {
                    console.error('Error parsing clarification data:', e);
                }
            }
            displayClarificationOptions(message.message, message.options, newCount);
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

    // Add user query to conversation history
    addToConversationHistory('user', query);

    // Show loading state
    submitBtn.disabled = true;
    submitBtn.textContent = 'Processing...';
    responseContainer.className = 'response-area';
    responseContent.innerHTML = '';
    clearMessages();
    clearStreamingStatus();    
    // Reset clarification count for new query
    currentClarificationCount = 0;
    // Check if streaming is enabled (v2 with WebSocket)
    const useStreaming = document.getElementById('useStreaming').checked;

    if (useStreaming && apiVersion === 'v2') {
        // Use SSE for streaming
        addStreamingStatus('Initializing streaming connection...', 'info');
        connectSSE(query, platform, model, previousConversation);
    } else {
        // Use traditional HTTP request
        addStreamingStatus('Sending request...', 'processing');
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

            const response = await fetch(`${getBasePath()}/handle-query-${apiVersion}?platform=${encodeURIComponent(platform)}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(requestBody)
            });

            const data = await response.json();

            if (response.ok) {
                addStreamingStatus('Response received', 'success');
                displayResponse(data.content.result || data.content);
            } else {
                addStreamingStatus('Error occurred', 'error');
                showError(data.message || 'An error occurred while processing your query');
            }
        } catch (error) {
            addStreamingStatus('Network error', 'error');
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

        // Add assistant response to conversation history
        addToConversationHistory('assistant', resultContent);

        // Clear the query input for the next question
        document.getElementById('query').value = '';

        // Add divider if there are streaming status items
        if (inlineStreamingContainer.children.length > 0) {
            responseContent.innerHTML = '<hr class="response-divider"><div class="final-response">' + renderMarkdown(resultContent) + '</div>';
        } else {
            responseContent.innerHTML = renderMarkdown(resultContent);
        }

        // Display follow-ups if available
        if (content.followUps && content.followUps.length > 0) {
            displayFollowUps(content.followUps);
        }
    } catch (error) {
        console.error('Error displaying response:', error);
        // If not JSON, display as markdown
        const errorContent = String(content);
        addToConversationHistory('assistant', errorContent);
        document.getElementById('query').value = '';

        if (inlineStreamingContainer.children.length > 0) {
            responseContent.innerHTML = '<hr class="response-divider"><div class="final-response">' + renderMarkdown(errorContent) + '</div>';
        } else {
            responseContent.innerHTML = renderMarkdown(errorContent);
        }
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
    followUpsContainer.addEventListener('click', function (e) {
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

// Display clarification options when the bot needs user input
function displayClarificationOptions(message, options, clarificationCount = 0) {
    responseContainer.className = 'response-area';

    // Store the clarification count for use when sending response
    currentClarificationCount = clarificationCount;

    const optionsHtml = options.map((option, index) => `
        <button class="clarification-option" data-option="${escapeHtml(option)}">
            ${escapeHtml(option)}
        </button>
    `).join('');

    responseContent.innerHTML = `
        <div class="clarification-container">
            <div class="clarification-message">
                <span class="clarification-icon">🤔</span>
                ${escapeHtml(message)}
            </div>
            <div class="clarification-options">
                ${optionsHtml}
            </div>
        </div>
    `;

    // Add click handlers for options
    const optionButtons = responseContent.querySelectorAll('.clarification-option');
    optionButtons.forEach(button => {
        button.addEventListener('click', function () {
            const selectedOption = this.getAttribute('data-option');
            sendClarificationResponse(selectedOption);
        });
    });
}

// Send user's clarification response back to the server
function sendClarificationResponse(selectedOption) {
    // Show selection confirmation
    responseContent.innerHTML = `
        <div class="clarification-selected">
            <span class="selected-icon">✓</span>
            <span class="selected-text"><strong>${escapeHtml(selectedOption)}</strong> selected</span>
        </div>
    `;
    addStreamingStatus('Processing your selection...', 'processing');

    // Make new SSE request with clarification and incremented count
    connectSSE(currentQuery, currentPlatform, currentModel, currentPreviousConversation, selectedOption, currentClarificationCount);
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
