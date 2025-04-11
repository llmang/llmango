package templates_templ

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
)

// RenderTempl renders a templ component to the ResponseWriter
func RenderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) error {
	// Set common headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render the component
	return component.Render(r.Context(), w)
}

// WithBaseRoute defines a context key type for base route
type contextKey string

const BaseRouteKey contextKey = "baseRoute"

// WithBaseRoute adds the base route to the context
func WithBaseRoute(ctx context.Context, baseRoute string) context.Context {
	return context.WithValue(ctx, BaseRouteKey, baseRoute)
}

// GetBaseRoute gets the base route from context
func GetBaseRoute(ctx context.Context) string {
	if baseRoute, ok := ctx.Value(BaseRouteKey).(string); ok {
		return baseRoute
	}
	return ""
}
