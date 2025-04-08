# LLMango Log Viewer Implementation

## Goals
1. Create a reusable log viewer component using Alpine.js and Go templates
2. Implement API endpoints for log retrieval with pagination
3. Display recent logs (5 most recent) for prompts and goals on their respective pages
4. Handle cases where logging is not enabled gracefully

## API Endpoints to Implement
1. General log endpoint with filter support
2. Prompt-specific log endpoint
3. Goal-specific log endpoint

## Data Structures

### Log Entry Structure
```json
{
  "timestamp": "string",
  "level": "string",
  "message": "string",
  "goal_uid": "string",
  "prompt_uid": "string",
  "metadata": {
    // Additional context-specific data
  }
}
```

### API Response Structure
```json
{
  "logs": [
    // Array of log entries
  ],
  "pagination": {
    "total": "number",
    "page": "number",
    "per_page": "number",
    "total_pages": "number"
  }
}
```

## Implementation Steps
1. Create logging.go template with reusable log viewer component
2. Implement API endpoints in api_router.go
3. Add log viewer sections to prompt and goal pages
4. Implement pagination on both frontend and backend
5. Add error handling for disabled logging

## Notes
- Use Alpine.js for frontend interactivity
- Implement proper pagination for efficient data loading
- Handle null logging gracefully with user-friendly messages
- Ensure consistent styling with existing UI 