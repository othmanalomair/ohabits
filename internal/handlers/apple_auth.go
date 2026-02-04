package handlers

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"ohabits/internal/database"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// AppleSignInRequest represents the request from iOS/macOS app
type AppleSignInRequest struct {
	IdentityToken     string `json:"identityToken"`
	AuthorizationCode string `json:"authorizationCode"`
	FullName          string `json:"fullName"`
	Email             string `json:"email"`
}

// AppleSignInResponse represents the response to the native app
type AppleSignInResponse struct {
	Token     string `json:"token"`
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	FullName  string `json:"fullName"`
	IsNewUser bool   `json:"isNewUser"`
	Role      int    `json:"role"`
}

// AppleIDTokenClaims represents the claims in Apple's identity token
type AppleIDTokenClaims struct {
	Issuer          string `json:"iss"`
	Audience        string `json:"aud"`
	ExpiresAt       int64  `json:"exp"`
	IssuedAt        int64  `json:"iat"`
	Subject         string `json:"sub"` // Apple's unique user ID
	Email           string `json:"email"`
	EmailVerified   interface{} `json:"email_verified"`
	IsPrivateEmail  interface{} `json:"is_private_email"`
	RealUserStatus  int    `json:"real_user_status"`
	TransferSub     string `json:"transfer_sub"`
	NonceSupported  bool   `json:"nonce_supported"`
	AuthTime        int64  `json:"auth_time"`
	jwt.RegisteredClaims
}

// ApplePublicKey represents a public key from Apple's JWKS
type ApplePublicKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// AppleJWKS represents Apple's JSON Web Key Set
type AppleJWKS struct {
	Keys []ApplePublicKey `json:"keys"`
}

// appleJWKSCache caches Apple's public keys
var appleJWKSCache *AppleJWKS
var appleJWKSCacheTime time.Time
var appleJWKSCacheDuration = 24 * time.Hour

// AppleSignIn handles Apple Sign-In verification and user creation
func (h *Handler) AppleSignIn(c echo.Context) error {
	var req AppleSignInRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	if req.IdentityToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Identity token is required",
		})
	}

	ctx := c.Request().Context()

	// Verify Apple's identity token
	claims, err := verifyAppleIdentityToken(req.IdentityToken)
	if err != nil {
		log.Printf("Apple Sign-In token verification failed: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": fmt.Sprintf("Invalid identity token: %v", err),
		})
	}

	// Get Apple's unique user ID
	appleUserID := claims.Subject
	if appleUserID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid identity token: missing subject",
		})
	}

	// Determine email - from claims or request (Apple only sends email on first sign-in)
	email := claims.Email
	if email == "" && req.Email != "" {
		email = req.Email
	}

	// Determine display name
	displayName := req.FullName
	if displayName == "" {
		// Use email prefix as fallback
		if email != "" {
			parts := strings.Split(email, "@")
			displayName = parts[0]
		} else {
			displayName = "مستخدم Apple"
		}
	}

	var user *database.User
	var isNewUser bool

	// Try to find user by Apple ID first
	user, err = h.DB.GetUserByAppleID(ctx, appleUserID)
	if err != nil {
		if err == database.ErrUserNotFound {
			// User doesn't exist with this Apple ID
			// Check if there's a user with this email to link
			if email != "" {
				existingUser, emailErr := h.DB.GetUserByEmail(ctx, email)
				if emailErr == nil && existingUser != nil {
					// Link Apple ID to existing user
					if linkErr := h.DB.LinkAppleID(ctx, existingUser.ID, appleUserID); linkErr != nil {
						return c.JSON(http.StatusInternalServerError, map[string]string{
							"error": "Failed to link Apple ID to existing account",
						})
					}
					user = existingUser
					isNewUser = false
				} else {
					// Create new user
					user, err = h.DB.CreateAppleUser(ctx, appleUserID, email, displayName)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, map[string]string{
							"error": fmt.Sprintf("Failed to create user: %v", err),
						})
					}
					isNewUser = true
				}
			} else {
				// No email, create user with a placeholder email
				placeholderEmail := fmt.Sprintf("%s@privaterelay.appleid.com", appleUserID)
				user, err = h.DB.CreateAppleUser(ctx, appleUserID, placeholderEmail, displayName)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{
						"error": fmt.Sprintf("Failed to create user: %v", err),
					})
				}
				isNewUser = true
			}
		} else {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Database error",
			})
		}
	}

	// Generate JWT token
	token, err := h.Auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate authentication token",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"fullName": user.DisplayName,
			"role":     user.Role,
		},
		"isNewUser": isNewUser,
	})
}

// verifyAppleIdentityToken verifies an Apple identity token
func verifyAppleIdentityToken(tokenString string) (*AppleIDTokenClaims, error) {
	// Parse token without verification first to get the key ID
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &AppleIDTokenClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Get key ID from header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("missing key ID in token header")
	}

	// Get Apple's public keys
	jwks, err := getApplePublicKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to get Apple public keys: %w", err)
	}

	// Find the matching key
	var publicKey *rsa.PublicKey
	for _, key := range jwks.Keys {
		if key.Kid == kid {
			publicKey, err = parseRSAPublicKey(key)
			if err != nil {
				return nil, fmt.Errorf("failed to parse public key: %w", err)
			}
			break
		}
	}

	if publicKey == nil {
		return nil, errors.New("no matching public key found")
	}

	// Verify the token
	claims := &AppleIDTokenClaims{}
	token, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Verify issuer
	if claims.Issuer != "https://appleid.apple.com" {
		return nil, errors.New("invalid issuer")
	}

	// Verify audience (bundle ID)
	if claims.Audience != "com.most3mr.ohabits" {
		return nil, fmt.Errorf("invalid audience: got %s", claims.Audience)
	}

	// Verify expiration
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// getApplePublicKeys fetches Apple's public keys from their JWKS endpoint
func getApplePublicKeys() (*AppleJWKS, error) {
	// Check cache
	if appleJWKSCache != nil && time.Since(appleJWKSCacheTime) < appleJWKSCacheDuration {
		return appleJWKSCache, nil
	}

	// Fetch from Apple
	resp, err := http.Get("https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Apple public keys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Apple JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks AppleJWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode Apple JWKS: %w", err)
	}

	// Update cache
	appleJWKSCache = &jwks
	appleJWKSCacheTime = time.Now()

	return &jwks, nil
}

// parseRSAPublicKey converts an Apple public key to an RSA public key
func parseRSAPublicKey(key ApplePublicKey) (*rsa.PublicKey, error) {
	// Decode N (modulus)
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	// Decode E (exponent)
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}
	// Convert exponent bytes to int
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}
