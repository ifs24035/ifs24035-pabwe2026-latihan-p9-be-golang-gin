package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"app/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.AbortFail(c, http.StatusUnauthorized, "Unauthorized: Missing Authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.AbortFail(c, http.StatusUnauthorized, "Unauthorized: Invalid Authorization header format")
			return
		}

		tokenString := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			response.AbortFail(c, http.StatusUnauthorized, "Unauthorized: Invalid or expired token")
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("userRole")
		if !exists {
			response.AbortFail(c, http.StatusForbidden, "Forbidden: User role not found")
			return
		}

		userRole := roleVal.(string)
		for _, r := range allowedRoles {
			if r == userRole {
				c.Next()
				return
			}
		}

		response.AbortFail(c, http.StatusForbidden, "Forbidden: Insufficient privileges")
	}
}

type rateBucket struct {
	tokens    int
	lastCheck time.Time
}

var (
	limiterMu      sync.Mutex
	limiterBuckets = make(map[string]*rateBucket)
)

func RateLimiter(maxRequests int, per time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiterMu.Lock()
		now := time.Now()
		bucket, exists := limiterBuckets[ip]
		if !exists || now.Sub(bucket.lastCheck) > per {
			limiterBuckets[ip] = &rateBucket{
				tokens:    maxRequests - 1,
				lastCheck: now,
			}
			limiterMu.Unlock()
			c.Next()
			return
		}

		if bucket.tokens > 0 {
			bucket.tokens--
			limiterMu.Unlock()
			c.Next()
			return
		}

		limiterMu.Unlock()
		response.AbortFail(c, http.StatusTooManyRequests, "Too many requests. Please slow down.")
	}
}
