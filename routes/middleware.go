package routes

import (
	"context"
	"fmt"
	"forum/internal/data"
	"forum/models"
	"forum/utils"
	"log"
	"net/http"
)

type ContextKey string // ContextKey types for different context values

const (
	DBManagerKey ContextKey = "dbManager"
	UserKey      ContextKey = "user"
	SessionKey   ContextKey = "session"
)

// Middleware represents a function that wraps an http.HandlerFunc
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Chain creates a middleware chain following daemon pattern
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Apply middlewares in reverse order to create proper chain
			handler := final
			for i := len(middlewares) - 1; i >= 0; i-- {
				handler = middlewares[i](handler)
			}
			handler(w, r)
		}
	}
}

// WithDatabaseManager middleware adds database manager to context
func WithDatabaseManager(dbManager *data.DatabaseManager) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), DBManagerKey, dbManager)
			next(w, r.WithContext(ctx))
		}
	}
}

// WithAuthentication middleware checks for session and adds user to context
func WithAuthentication() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			dbManager := GetDatabaseManager(r)
			if dbManager == nil {
				next(w, r)
				return
			}

			// Check for session cookie
			cookie, err := r.Cookie("_cookie")
			if err != nil {
				// No cookie, continue without authentication
				next(w, r)
				return
			}

			// Validate session
			session, err := dbManager.GetSessionByCookie(cookie.Value)
			if err != nil {
				// Invalid session, continue without authentication
				next(w, r)
				return
			}

			// Validate and update session activity
			_, isValid, err := dbManager.ValidateSession(session.Uuid)
			if err != nil || !isValid {
				// Invalid session, continue without authentication
				next(w, r)
				return
			}

			// Get user from session
			user, err := dbManager.GetUserByID(session.UserId)
			if err != nil {
				next(w, r) // User not found, continue without authentication
				return
			}
			// Add session and user to context
			ctx := context.WithValue(r.Context(), SessionKey, session)
			ctx = context.WithValue(ctx, UserKey, user)
			next(w, r.WithContext(ctx)) // Continue to next handler with updated context
		}
	}
}

// RequireAuth middleware ensures user is authenticated
func RequireAuth() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user := GetCurrentUser(r)
			if user == nil {
				http.Redirect(w, r, "/login/", http.StatusFound)
				return
			}
			next(w, r)
		}
	}
}

// WithLogging middleware logs requests
func WithLogging() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("[%s] %s %s\n", r.Method, r.URL.Path, r.RemoteAddr)
			next(w, r)
		}
	}
}

// WithErrorRecovery middleware handles panics and converts them to 500 errors
func WithErrorRecovery() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Panic recovered: %v", err)
					utils.InternalServerError(w, r, fmt.Errorf("panic: %v", err))
				}
			}()
			next(w, r)
		}
	}
}

// WithNotFoundHandler creates a 404 handler
func NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.NotFound(w, r)
	}
}

// Helper functions to retrieve values from context

// GetDatabaseManager retrieves database manager from context
func GetDatabaseManager(r *http.Request) *data.DatabaseManager {
	if dbManager, ok := r.Context().Value(DBManagerKey).(*data.DatabaseManager); ok {
		return dbManager
	}
	return nil
}

// GetCurrentUser retrieves authenticated user from context
func GetCurrentUser(r *http.Request) *models.User {
	if user, ok := r.Context().Value(UserKey).(models.User); ok {
		return &user
	}
	return nil
}

// GetCurrentSession retrieves session from context
func GetCurrentSession(r *http.Request) *models.Session {
	if session, ok := r.Context().Value(SessionKey).(models.Session); ok {
		return &session
	}
	return nil
}

// IsAuthenticated checks if user is authenticated
func IsAuthenticated(r *http.Request) bool {
	return GetCurrentUser(r) != nil
}

// SessionCheck checks if user has valid session (for backward compatibility)
func SessionCheck(w http.ResponseWriter, r *http.Request) (*models.Session, error) {
	session := GetCurrentSession(r)
	if session == nil {
		return nil, fmt.Errorf("no valid session")
	}
	return session, nil
}
