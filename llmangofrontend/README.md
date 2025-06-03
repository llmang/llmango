# LLMango Frontend

Web-based UI for managing Goals and Prompts with real-time monitoring and testing capabilities.

## Features

### Goal & Prompt Management ✅
- Create, edit, and delete goals and prompts through web interface
- Support for both typed and JSON-based goal creation
- Real-time validation and testing

### Monitoring & Analytics ✅
- Live execution logs with filtering and search
- Goal performance metrics and success rates
- Prompt A/B testing results and analytics

### API Integration ✅
RESTful API for frontend-backend communication:

```go
// Goal management
GET/POST/PUT/DELETE /api/goals
GET/POST/PUT/DELETE /api/prompts

// Execution and monitoring  
GET /api/logs
POST /api/execute
GET /api/analytics
```

### Svelte Frontend ✅
Modern, responsive web interface built with SvelteKit:

- [`svelte/src/routes/`](svelte/src/routes/) - Page components
- [`svelte/src/lib/`](svelte/src/lib/) - Reusable UI components
- [`svelte/src/lib/classes/`](svelte/src/lib/classes/) - API client classes

## Key Components

### Backend
- [`router.go`](router.go) - Main HTTP router and middleware
- [`api_goals.go`](api_goals.go) - Goal management endpoints
- [`api_prompts.go`](api_prompts.go) - Prompt management endpoints
- [`api_logs.go`](api_logs.go) - Logging and monitoring endpoints

### Frontend
- [`svelte/src/routes/goal/`](svelte/src/routes/goal/) - Goal management pages
- [`svelte/src/routes/prompt/`](svelte/src/routes/prompt/) - Prompt management pages
- [`svelte/src/routes/logs/`](svelte/src/routes/logs/) - Monitoring dashboard

## Setup

Mount the frontend at `/mango` with proper authentication:

```go
mangoRouter := llmangofrontend.CreateLLMMangRouter(llmangoManager, nil)
router.Handle("/mango/", http.StripPrefix("/mango", middleware.Auth(mangoRouter)))
```

## Status: ✅ Complete

Full-featured web interface with goal management, real-time monitoring, and comprehensive analytics.