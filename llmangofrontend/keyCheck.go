package llmangofrontend

import (
	"fmt"
	"net/http"
	"strings"
)

// Middleware function to check for API key
func (r *Router) apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Allow requests to update-key without checking the key
		if req.URL.Path == r.BaseRoute+"/api/update-key" {
			next.ServeHTTP(w, req)
			return
		}

		// Check if the API key is present
		if r.LLMangoManager.OpenRouter == nil || r.LLMangoManager.OpenRouter.ApiKey == "" {
			// API key is missing, serve the page to enter the key
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized) // Use Unauthorized status

			// Replace the API URL placeholder with the actual base route
			htmlWithRoute := strings.Replace(
				apiKeyPageHTML,
				"{baseRoute}",
				r.BaseRoute,
				-1,
			)

			fmt.Fprint(w, htmlWithRoute)
			return // Stop processing the request further
		}

		// API key exists, proceed to the next handler
		next.ServeHTTP(w, req)
	})
}

const apiKeyPageHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Enter API Key</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
    <h1>Enter OpenRouter API Key</h1>
    <form id="apiKeyForm">
        <label for="apiKey">API Key:</label>
        <input type="text" id="apiKey" name="apiKey" required size="50">
        <button type="submit">Update Key</button>
    </form>
    <p id="message"></p>
	<button id="refreshButton" style="display:none;" onclick="location.reload()">Refresh Page</button>

    <script>
        // Set the API base route from the server
        const baseRoute = "{baseRoute}";
        
        document.getElementById('apiKeyForm').addEventListener('submit', function(event) {
            event.preventDefault(); // Prevent default form submission
            const apiKey = document.getElementById('apiKey').value;
            const messageElement = document.getElementById('message');
			const refreshButton = document.getElementById('refreshButton');

            messageElement.textContent = 'Updating...';

            fetch(baseRoute + '/api/update-key', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ apiKey: apiKey }) // Match the expected structure
            })
            .then(response => {
                if (!response.ok) {
                    return response.text().then(text => { throw new Error('Failed to update key: ' + text); });
                }
                return response.text();
            })
            .then(data => {
                messageElement.textContent = data; // Display success message
				refreshButton.style.display = 'block';
                // Optionally clear the input field
                // document.getElementById('apiKey').value = '';
            })
            .catch(error => {
                console.error('Error:', error);
                messageElement.textContent = 'Error: ' + error.message;
				refreshButton.style.display = 'none';
            });
        });
    </script>
</body>
</html>
`
