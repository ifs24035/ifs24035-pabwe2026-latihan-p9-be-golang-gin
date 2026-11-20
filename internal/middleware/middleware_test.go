package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func generateTestToken(secret string, userID uint, username, role string, expired bool) string {
	expiration := time.Hour
	if expired {
		expiration = -time.Hour
	}
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func TestCORS(t *testing.T) {
	r := gin.New()
	r.Use(CORS())
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Test OPTIONS preflight
	req, _ := http.NewRequest(http.MethodOptions, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

	// Test GET normal request
	reqGet, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)
}

func TestAuth(t *testing.T) {
	secret := "test-secret"
	r := gin.New()
	r.Use(Auth(secret))
	r.GET("/protected", func(c *gin.Context) {
		uid, _ := c.Get("userID")
		uname, _ := c.Get("username")
		role, _ := c.Get("userRole")
		c.JSON(http.StatusOK, gin.H{
			"user_id":  uid,
			"username": uname,
			"role":     role,
		})
	})

	// 1. Missing header
	req1, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusUnauthorized, w1.Code)

	// 2. Invalid format
	req2, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "InvalidFormatToken")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	// 3. Expired token
	expiredToken := generateTestToken(secret, 1, "testuser", "member", true)
	req3, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req3.Header.Set("Authorization", "Bearer "+expiredToken)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusUnauthorized, w3.Code)

	// 4. Valid token
	validToken := generateTestToken(secret, 42, "tester", "member", false)
	req4, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req4.Header.Set("Authorization", "Bearer "+validToken)
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)
}

func TestRequireRole(t *testing.T) {
	r := gin.New()
	r.GET("/admin-only", func(c *gin.Context) {
		c.Set("userRole", "admin")
	}, RequireRole("admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "admin granted")
	})

	r.GET("/member-only", func(c *gin.Context) {
		c.Set("userRole", "member")
	}, RequireRole("admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "should be forbidden")
	})

	r.GET("/no-role", RequireRole("admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "should be forbidden")
	})

	// Granted
	req1, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Forbidden
	req2, _ := http.NewRequest(http.MethodGet, "/member-only", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusForbidden, w2.Code)

	// No role in context
	req3, _ := http.NewRequest(http.MethodGet, "/no-role", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusForbidden, w3.Code)
}

func TestRateLimiter(t *testing.T) {
	r := gin.New()
	r.Use(RateLimiter(2, time.Second))
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// 1st request - ok
	req1, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 2nd request - ok
	req2, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// 3rd request - throttled (TooManyRequests)
	req3, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)

	// Reset buckets after duration
	time.Sleep(1100 * time.Millisecond)
	req4, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)
}
