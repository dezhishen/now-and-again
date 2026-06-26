package middleware

import (
	"net/http"
	"strings"

	"github.com/dezhishen/now-and-again/shared/scopes"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CORS enables cross-origin requests with credentials support for cookies.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-API-Key")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// ApiKeyValidator validates API keys and returns scopes.
type ApiKeyValidator interface {
	ValidateApiKey(raw string) (userID string, scopes []string, err error)
}

// JWTAuth validates the Bearer token (JWT or API Key).
func JWTAuth(secret string, apiKeyValidator ApiKeyValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Authorization header first, then X-API-Key
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			authHeader = c.GetHeader("X-API-Key")
		}

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		tokenStr := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Try API Key first (prefix "na_")
		if strings.HasPrefix(tokenStr, "na_") && apiKeyValidator != nil {
			userID, scopes, err := apiKeyValidator.ValidateApiKey(tokenStr)
			if err == nil && userID != "" {
				c.Set("user_id", userID)
				c.Set("auth_method", "api_key")
				c.Set("api_key_scopes", scopes)
				c.Next()
				return
			}
		}

		// Try JWT
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}

		userID, _ := claims["sub"].(string)
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing subject"})
			return
		}

		c.Set("user_id", userID)
		c.Set("auth_method", "jwt")
		c.Next()
	}
}

// ScopeGuard returns a middleware that checks the required scope based on the
// request's HTTP method and path. JWT users always pass; API key users must have
// the scope defined in scopes.RouteScope for the current route.
func ScopeGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		authMethod, _ := c.Get("auth_method")
		if authMethod != "api_key" {
			c.Next()
			return
		}
		required := scopes.RouteScope(c.Request.Method, c.FullPath())
		if required == "" {
			// No scope defined for this route — allow (public routes are not in auth group)
			c.Next()
			return
		}
		granted, _ := c.Get("api_key_scopes")
		grantedList, _ := granted.([]string)
		if !scopes.Has(grantedList, required) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient scope, need: " + required})
			return
		}
		c.Next()
	}
}
