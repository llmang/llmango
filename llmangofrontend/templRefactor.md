# Go Template to Templ Conversion Plan

## Overview of Differences

| Feature | Go Templates | Templ |
|---------|-------------|-------|
| File Format | Embedded strings or external files | `.templ` files with Go-like syntax |
| Type Safety | Limited type checking | Strong type safety with Go code generation |
| Reusability | Template inheritance with `{{define}}` and `{{template}}` | Components as Go functions |
| Output | Plain strings via `template.Execute()` | Components implementing `templ.Component` interface |
| Expressions | `{{.Field}}` for data access | `{ expression }` for Go expressions |
| Logic | `{{if}}`, `{{range}}`, etc. | Native Go `if`, `for`, `switch` statements |
| Function Map | Required for custom functions | Native Go functions |

## Current Structure Analysis

The current template system:
- Uses Go's `html/template` package
- Templates are defined as string constants in the `templates` package
- Templates are parsed and stored in a global `Templates` variable
- Uses a custom `renderTemplate` function to execute templates
- Employs custom helper functions via `template.FuncMap` (e.g., `dict`, `toJSON`, `getGoalInfo`)

## Conversion Steps

### 1. Project Setup

1. Install templ CLI:
   ```
   go install github.com/a-h/templ/cmd/templ@latest
   ```

2. Update module dependencies:
   ```
   go get github.com/a-h/templ
   ```

3. Create templates directory structure:
   ```
   llmangofrontend/templates_templ/
   ```

4. Add templ generation to build process

### 2. Component Conversion Strategy

1. **Start with shared components**: Convert shared.go first to establish patterns
2. **Convert page templates**: Move from simple to complex pages
3. **Maintain parallel systems**: Keep Go templates working while incrementally adding templ components
4. **Update routing**: Modify handlers to use new templ components

### 3. Converting Individual Templates

For each template:

1. Create a corresponding `.templ` file
2. Convert Go template syntax to templ syntax:
   - `{{.Variable}}` → `{ variable }`
   - `{{template "name" .}}` → `@nameComponent(data)`
   - `{{range .Items}}` → `for _, item := range items {}`
   - `{{if .Condition}}` → `if condition {}`
   - Custom template functions → native Go functions

3. Organize components hierarchically:
   - Create base layout components
   - Split into reusable components
   - Define page-specific components

### 4. Converting Custom Functions

1. Replace template functions with Go functions:
   - `dict` → Pass structured data to components
   - `toJSON` → Use Go's json package directly
   - `inc` → Use native Go arithmetic
   - `getGoalInfo` → Create dedicated Go functions

### 5. Rendering Implementation

Create new render functions:

```go
// Replace renderTemplate with templ component rendering
func (r *Router) renderTempl(w http.ResponseWriter, component templ.Component) {
    component.Render(req.Context(), w)
}
```

### 6. Router Update

Update HTTP handlers to use templ components:

```go
// Before
func (r *Router) handleHomePage(w http.ResponseWriter, req *http.Request) {
    data := HomePageData{
        Prompts: r.Prompts,
        Goals:   r.Goals,
    }
    r.renderTemplate(w, "home", data)
}

// After
func (r *Router) handleHomePage(w http.ResponseWriter, req *http.Request) {
    component := pages.HomePage(r.Prompts, r.Goals)
    component.Render(req.Context(), w)
}
```

## Common Issues and Solutions

### 1. Import Conflicts

**Problem:** Multiple imports of the templ package can cause redeclaration errors.

**Solution:** 
- Create utility files (e.g., `utils.templ`) for shared functions like `SafeURL`
- Import the templ package only once in these utility files
- Use the functions from other templates without importing templ again

### 2. Variable Scope Issues

**Problem:** Variables declared in template loops may not be accessible in the generated Go code.

**Solution:**
- Declare variables before loops or conditionals
- Use simplified logic where possible
- Avoid relying on declared variables across component boundaries

### 3. Type Safety with URLs

**Problem:** String URLs must be converted to `templ.SafeURL` for href attributes.

**Solution:**
- Create a helper function `SafeURL(url string) templ.SafeURL`
- Use this for all URL attributes: `<a href={ SafeURL("/path") }>`

### 4. Component Composition

**Problem:** The `dict` function in Go templates isn't needed in templ.

**Solution:**
- Pass individual variables to components instead of maps
- For complex data, create structs to represent the data

### 5. Error Handling

**Problem:** Error handling in templ is more explicit.

**Solution:**
- Handle errors from component rendering explicitly
- Use context for passing request-scoped values

### 7. Sample Conversion Examples

#### Go Template:
```html
{{define "home"}}
{{template "header" .}}
    <div class="section-header">
        <h2>Recent Goals</h2>
        <a href="{{.BaseRoute}}/goals">View All</a>
    </div>
    <div class="card-container">
        {{range $id, $goalAny := .Goals}}
            {{$goal := getGoalInfo $goalAny}}
            {{template "goal-card" dict "ID" $id "Goal" $goal "BaseRoute" $.BaseRoute}}
        {{end}}
    </div>
{{end}}
```

#### Templ Equivalent:
```templ
package pages

import "github.com/llmang/llmango/llmango"

templ HomePage(prompts map[string]*llmango.Prompt, goals map[string]any, baseRoute string) {
    @Layout() {
        <div class="section-header">
            <h2>Recent Goals</h2>
            <a href={ SafeURL(baseRoute + "/goals") }>View All</a>
        </div>
        <div class="card-container">
            for id, goalAny := range goals {
                @GoalCard(id, getGoalInfo(goalAny), baseRoute)
            }
        </div>
    }
}
```

### 8. Testing Strategy

1. Start with isolated components
2. Create parallel routes for templ-based pages
3. Write unit tests for components
4. Test rendering against expected HTML output
5. A/B test both implementations before full migration

### 9. Refactoring Timeline

1. **Phase 1**: Create basic shared components (1-2 days)
2. **Phase 2**: Convert simple pages (2-3 days)
3. **Phase 3**: Convert complex pages with interactions (2-3 days)
4. **Phase 4**: Update all routes and handlers (1 day)
5. **Phase 5**: Testing and bug fixing (2-3 days)
6. **Phase 6**: Remove old template system (1 day)

## Benefits of Migration

1. **Type Safety**: Catch errors at compile time
2. **IDE Support**: Better autocompletion and refactoring tools
3. **Modularity**: More reusable and maintainable components
4. **Performance**: Potentially better rendering performance
5. **Developer Experience**: More Go-like syntax and patterns

## Potential Challenges

1. Learning curve for templ syntax
2. Handling complex custom helper functions
3. Maintaining backward compatibility during transition
4. Ensuring visual consistency between implementations
5. Managing template generation in build process 