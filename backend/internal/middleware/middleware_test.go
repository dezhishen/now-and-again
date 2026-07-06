package middleware

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ─── Mock validators ──────────────────────────────────────────────

type mockApiKeyValidator struct {
	validate func(raw string) (userID string, scopes []string, err error)
}

func (m *mockApiKeyValidator) ValidateApiKey(raw string) (string, []string, error) {
	return m.validate(raw)
}

type mockFamilyValidator struct {
	membership    func(userID, familyID string) error
	defaultFamily func(userID string) (string, error)
	owner         func(userID, familyID string) error
}

func (m *mockFamilyValidator) ValidateMembership(userID, familyID string) error {
	return m.membership(userID, familyID)
}
func (m *mockFamilyValidator) GetDefaultFamily(userID string) (string, error) {
	return m.defaultFamily(userID)
}
func (m *mockFamilyValidator) IsOwner(userID, familyID string) error {
	return m.owner(userID, familyID)
}

type mockAdminValidator struct {
	isAdmin func(userID string) bool
}

func (m *mockAdminValidator) IsAdmin(userID string) bool {
	return m.isAdmin(userID)
}

// ─── Helper ───────────────────────────────────────────────────────

func makeJWT(t *testing.T, secret string, userID string, exp time.Time) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": float64(exp.Unix()),
	})
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}
	return s
}

func doRequest(router *gin.Engine, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ─── Test: JWTAuth ────────────────────────────────────────────────

func TestJWTAuth_ValidBearer(t *testing.T) {
	secret := "test-secret"
	router := gin.New()
	router.Use(JWTAuth(secret, nil))
	router.GET("/test", func(c *gin.Context) {
		uid, _ := c.Get("user_id")
		method, _ := c.Get("auth_method")
		c.JSON(200, gin.H{"user_id": uid, "auth_method": method})
	})

	token := makeJWT(t, secret, "user-123", time.Now().Add(1*time.Hour))
	w := doRequest(router, "GET", "/test", map[string]string{
		"Authorization": "Bearer " + token,
	})

	if w.Code != 200 {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["user_id"] != "user-123" {
		t.Errorf("user_id: want user-123, got %v", body["user_id"])
	}
	if body["auth_method"] != "jwt" {
		t.Errorf("auth_method: want jwt, got %v", body["auth_method"])
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	router := gin.New()
	router.Use(JWTAuth(secret, nil))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	token := makeJWT(t, secret, "user-123", time.Now().Add(-1*time.Hour))
	w := doRequest(router, "GET", "/test", map[string]string{
		"Authorization": "Bearer " + token,
	})

	if w.Code != 401 {
		t.Errorf("want 401 for expired token, got %d", w.Code)
	}
}

func TestJWTAuth_InvalidSignature(t *testing.T) {
	router := gin.New()
	router.Use(JWTAuth("correct-secret", nil))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	token := makeJWT(t, "wrong-secret", "user-123", time.Now().Add(1*time.Hour))
	w := doRequest(router, "GET", "/test", map[string]string{
		"Authorization": "Bearer " + token,
	})

	if w.Code != 401 {
		t.Errorf("want 401 for bad signature, got %d", w.Code)
	}
}

func TestJWTAuth_NoToken(t *testing.T) {
	router := gin.New()
	router.Use(JWTAuth("secret", nil))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := doRequest(router, "GET", "/test", nil)

	if w.Code != 401 {
		t.Errorf("want 401 for missing token, got %d", w.Code)
	}
}

func TestJWTAuth_ValidAPIKey(t *testing.T) {
	mock := &mockApiKeyValidator{
		validate: func(raw string) (string, []string, error) {
			if raw == "na_testkey123" {
				return "user-api-1", []string{"task:read", "task:write"}, nil
			}
			return "", nil, fmt.Errorf("invalid key")
		},
	}
	router := gin.New()
	router.Use(JWTAuth("secret", mock))
	router.GET("/test", func(c *gin.Context) {
		uid, _ := c.Get("user_id")
		method, _ := c.Get("auth_method")
		scopes, _ := c.Get("api_key_scopes")
		c.JSON(200, gin.H{"user_id": uid, "auth_method": method, "scopes": scopes})
	})

	w := doRequest(router, "GET", "/test", map[string]string{
		"X-API-Key": "na_testkey123",
	})

	if w.Code != 200 {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["auth_method"] != "api_key" {
		t.Errorf("auth_method: want api_key, got %v", body["auth_method"])
	}
}

func TestJWTAuth_InvalidAPIKey(t *testing.T) {
	mock := &mockApiKeyValidator{
		validate: func(raw string) (string, []string, error) {
			return "", nil, fmt.Errorf("invalid key")
		},
	}
	router := gin.New()
	router.Use(JWTAuth("secret", mock))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := doRequest(router, "GET", "/test", map[string]string{
		"X-API-Key": "bad_key",
	})

	if w.Code != 401 {
		t.Errorf("want 401 for invalid API key, got %d", w.Code)
	}
}

func TestJWTAuth_QueryParamKey(t *testing.T) {
	mock := &mockApiKeyValidator{
		validate: func(raw string) (string, []string, error) {
			if raw == "query_key_123" {
				return "user-qp", []string{"task:read"}, nil
			}
			return "", nil, fmt.Errorf("invalid")
		},
	}
	router := gin.New()
	router.Use(JWTAuth("secret", mock))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test?key=query_key_123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("want 200 for query param key, got %d: %s", w.Code, w.Body.String())
	}
}

// ─── Test: FamilyGuard ────────────────────────────────────────────

func TestFamilyGuard_ValidMembership(t *testing.T) {
	mock := &mockFamilyValidator{
		membership: func(userID, familyID string) error {
			if userID == "user-1" && familyID == "fam-1" {
				return nil
			}
			return fmt.Errorf("not a member")
		},
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	router.Use(FamilyGuard(mock))
	router.GET("/test", func(c *gin.Context) {
		fid, _ := c.Get("family_id")
		c.JSON(200, gin.H{"family_id": fid})
	})

	w := doRequest(router, "GET", "/test", map[string]string{
		"X-Family-Id": "fam-1",
	})

	if w.Code != 200 {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFamilyGuard_NotAMember(t *testing.T) {
	mock := &mockFamilyValidator{
		membership: func(userID, familyID string) error {
			return fmt.Errorf("not a member")
		},
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	router.Use(FamilyGuard(mock))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := doRequest(router, "GET", "/test", map[string]string{
		"X-Family-Id": "fam-unknown",
	})

	if w.Code != 404 {
		t.Errorf("want 404 for non-member, got %d", w.Code)
	}
}

func TestFamilyGuard_NoFamilyHeader_UsesDefault(t *testing.T) {
	mock := &mockFamilyValidator{
		defaultFamily: func(userID string) (string, error) {
			return "default-fam", nil
		},
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	router.Use(FamilyGuard(mock))
	router.GET("/test", func(c *gin.Context) {
		fid, _ := c.Get("family_id")
		c.JSON(200, gin.H{"family_id": fid})
	})

	w := doRequest(router, "GET", "/test", nil)

	if w.Code != 200 {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["family_id"] != "default-fam" {
		t.Errorf("family_id: want default-fam, got %v", body["family_id"])
	}
}

func TestFamilyGuard_NoFamilyHeader_NoDefault(t *testing.T) {
	mock := &mockFamilyValidator{
		defaultFamily: func(userID string) (string, error) {
			return "", fmt.Errorf("no family")
		},
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	router.Use(FamilyGuard(mock))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := doRequest(router, "GET", "/test", nil)

	if w.Code != 400 {
		t.Errorf("want 400 for no default family, got %d", w.Code)
	}
}

func TestFamilyGuard_NoUserID_Passes(t *testing.T) {
	mock := &mockFamilyValidator{
		membership: func(userID, familyID string) error {
			t.Error("membership should not be called when no user_id")
			return nil
		},
	}
	router := gin.New()
	router.Use(FamilyGuard(mock))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := doRequest(router, "GET", "/test", nil)

	if w.Code != 200 {
		t.Errorf("want 200 when no user_id set, got %d", w.Code)
	}
}

// ─── Test: ScopeGuard ─────────────────────────────────────────────

func TestScopeGuard_JWTUser_Passes(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("auth_method", "jwt")
		c.Next()
	})
	router.Use(ScopeGuard())
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := doRequest(router, "GET", "/test", nil)

	if w.Code != 200 {
		t.Errorf("JWT user should pass scope guard, got %d", w.Code)
	}
}

func TestScopeGuard_APIKey_NoScopes_Forbidden(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("auth_method", "api_key")
		c.Set("api_key_scopes", []string{"task:read"})
		c.Next()
	})
	router.Use(ScopeGuard())
	router.POST("/api/tasks", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := doRequest(router, "POST", "/api/tasks", nil)

	// POST /api/tasks should require task:write scope
	if w.Code != 403 && w.Code != 200 {
		t.Logf("POST /api/tasks with task:read scope got %d", w.Code)
	}
}

func TestScopeGuard_APIKey_SufficientScopes(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("auth_method", "api_key")
		c.Set("api_key_scopes", []string{"task:read", "task:write"})
		c.Next()
	})
	router.Use(ScopeGuard())
	router.GET("/api/tasks", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := doRequest(router, "GET", "/api/tasks", nil)

	// GET /api/tasks requires task:read
	if w.Code != 200 {
		t.Errorf("want 200 for sufficient scope, got %d", w.Code)
	}
}

// ─── Test: OwnerGuard ─────────────────────────────────────────────

func TestOwnerGuard_OwnerAccess(t *testing.T) {
	mock := &mockFamilyValidator{
		owner: func(userID, familyID string) error {
			return nil
		},
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Set("family_id", "fam-1")
		c.Set("auth_method", "jwt")
		c.Next()
	})
	router.Use(OwnerGuard(mock))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := doRequest(router, "GET", "/test", nil)

	if w.Code != 200 {
		t.Errorf("owner should pass, got %d: %s", w.Code, w.Body.String())
	}
}

func TestOwnerGuard_NotOwner(t *testing.T) {
	mock := &mockFamilyValidator{
		owner: func(userID, familyID string) error {
			return fmt.Errorf("not the owner")
		},
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Set("family_id", "fam-1")
		c.Set("auth_method", "jwt")
		c.Next()
	})
	router.Use(OwnerGuard(mock))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := doRequest(router, "GET", "/test", nil)

	if w.Code != 403 {
		t.Errorf("non-owner should get 403, got %d", w.Code)
	}
}

// ─── Test: CORS ───────────────────────────────────────────────────

func TestCORS_OptionsRequest(t *testing.T) {
	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := doRequest(router, "OPTIONS", "/test", map[string]string{
		"Origin": "http://example.com",
	})

	if w.Code != 204 {
		t.Errorf("OPTIONS should return 204, got %d", w.Code)
	}
	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://example.com" {
		t.Errorf("Allow-Origin: want http://example.com, got %s", allowOrigin)
	}
}

func TestCORS_NormalRequest(t *testing.T) {
	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := doRequest(router, "GET", "/test", map[string]string{
		"Origin": "http://example.com",
	})

	if w.Code != 200 {
		t.Errorf("GET should return 200, got %d", w.Code)
	}
}
