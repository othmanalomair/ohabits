package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

type AuthMiddleware struct {
	Secret []byte
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{
		Secret: []byte(secret),
	}
}

// GenerateToken creates a new JWT token for a user
func (m *AuthMiddleware) GenerateToken(userID uuid.UUID, email string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // 30 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.Secret)
}

// ValidateToken validates a JWT token and returns the claims
func (m *AuthMiddleware) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return m.Secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// RequireAuth middleware checks for valid JWT token
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Try cookie first
		cookie, err := c.Cookie("token")
		var tokenString string

		if err == nil {
			tokenString = cookie.Value
		} else {
			// Try Authorization header
			auth := c.Request().Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				tokenString = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if tokenString == "" {
			// Redirect to login for web requests
			if strings.Contains(c.Request().Header.Get("Accept"), "text/html") {
				return c.Redirect(http.StatusSeeOther, "/login")
			}
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "غير مصرح"})
		}

		claims, err := m.ValidateToken(tokenString)
		if err != nil {
			// Clear invalid cookie
			c.SetCookie(&http.Cookie{
				Name:     "token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
			})

			if strings.Contains(c.Request().Header.Get("Accept"), "text/html") {
				return c.Redirect(http.StatusSeeOther, "/login")
			}
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "جلسة منتهية"})
		}

		// Store user info in context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)

		return next(c)
	}
}

// OptionalAuth middleware extracts user info if available but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("token")
		if err == nil {
			claims, err := m.ValidateToken(cookie.Value)
			if err == nil {
				c.Set("userID", claims.UserID)
				c.Set("email", claims.Email)
			}
		}
		return next(c)
	}
}

// GetUserID extracts user ID from context
func GetUserID(c echo.Context) (uuid.UUID, bool) {
	userID, ok := c.Get("userID").(uuid.UUID)
	return userID, ok
}

// SetAuthCookie sets the authentication cookie
func SetAuthCookie(c echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearAuthCookie clears the authentication cookie
func ClearAuthCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
