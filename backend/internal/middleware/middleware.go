package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/scopes"
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
		if origin != "*" {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-API-Key,X-Family-Id")
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

// FamilyValidator checks family membership and returns the user's default family.
type FamilyValidator interface {
	ValidateMembership(userID, familyID string) error
	GetDefaultFamily(userID string) (string, error)
	IsOwner(userID, familyID string) error
}

// AdminValidator checks whether a user has the admin role.
type AdminValidator interface {
	IsAdmin(userID string) bool
}

// JWTAuth validates the Bearer token (JWT or API Key).
func JWTAuth(secret string, apiKeyValidator ApiKeyValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Authorization header first, then X-API-Key, then ?key= query param
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			authHeader = c.GetHeader("X-API-Key")
		}
		if authHeader == "" {
			authHeader = c.Query("key")
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
		}, jwt.WithLeeway(30*time.Second))
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}

		// Explicit exp check (defense-in-depth)
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
				return
			}
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

// FamilyGuard extracts the active family from the X-Family-Id header,
// validates membership, and sets "family_id" in the Gin context.
// Falls back to the user's default family if no header is provided.
// Archived families are only accessible by their owner.
func FamilyGuard(fv FamilyValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		uid, _ := userID.(string)
		if uid == "" {
			c.Next()
			return
		}

		familyID := c.GetHeader("X-Family-Id")

		if familyID != "" {
			if err := fv.ValidateMembership(uid, familyID); err != nil {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "family not found or not a member"})
				return
			}
		} else {
			defaultID, err := fv.GetDefaultFamily(uid)
			if err != nil || defaultID == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no family selected"})
				return
			}
			familyID = defaultID
		}

		c.Set("family_id", familyID)
		c.Next()
	}
}

// AdminGuard ensures the request comes from an admin user.
// JWT users must have the admin role; API-key users must have the admin:read scope.
func AdminGuard(av AdminValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := c.Get("user_id")
		userID, _ := uid.(string)
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		authMethod, _ := c.Get("auth_method")
		if authMethod == "api_key" {
			granted, _ := c.Get("api_key_scopes")
			grantedList, _ := granted.([]string)
			if !scopes.Has(grantedList, scopes.AdminRead) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin scope required"})
				return
			}
		} else {
			if av == nil || !av.IsAdmin(userID) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin role required"})
				return
			}
		}
		c.Next()
	}
}

// OwnerGuard ensures the requesting user is the owner of the active family.
func OwnerGuard(fv FamilyValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := c.Get("user_id")
		userID, _ := uid.(string)
		fid, _ := c.Get("family_id")
		familyID, _ := fid.(string)
		if userID == "" || familyID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		authMethod, _ := c.Get("auth_method")
		if authMethod == "api_key" {
			granted, _ := c.Get("api_key_scopes")
			grantedList, _ := granted.([]string)
			if !scopes.Has(grantedList, scopes.FamilyAdmin) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "family admin scope required"})
				return
			}
		} else {
			if err := fv.IsOwner(userID, familyID); err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "family owner required"})
				return
			}
		}
		c.Next()
	}
}
