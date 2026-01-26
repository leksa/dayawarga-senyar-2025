package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrNoAuthHeader     = errors.New("authorization header required")
	ErrInvalidAuthHeader = errors.New("invalid authorization header format")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrInvalidIssuer     = errors.New("invalid token issuer")
	ErrInvalidAudience   = errors.New("invalid token audience")
)

// OIDCConfig holds OIDC provider configuration
type OIDCConfig struct {
	IssuerURL string // e.g., http://localhost:9000/application/o/admin-portal/
	ClientID  string // The client ID (audience)
}

// OIDCClaims represents the JWT claims from Authentik
type OIDCClaims struct {
	jwt.RegisteredClaims
	Email         string   `json:"email,omitempty"`
	EmailVerified bool     `json:"email_verified,omitempty"`
	Name          string   `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Groups        []string `json:"groups,omitempty"`
	AvatarURL     string   `json:"avatar,omitempty"`
}

// JWKSet represents a JSON Web Key Set
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// OIDCValidator validates OIDC tokens from Authentik
type OIDCValidator struct {
	config     OIDCConfig
	httpClient *http.Client
	jwksURL    string

	// Cached JWKS
	mu         sync.RWMutex
	jwks       *JWKSet
	jwksExpiry time.Time
}

// NewOIDCValidator creates a new OIDC token validator
func NewOIDCValidator(config OIDCConfig) *OIDCValidator {
	// Normalize issuer URL (remove trailing slash)
	issuer := strings.TrimSuffix(config.IssuerURL, "/")

	return &OIDCValidator{
		config: OIDCConfig{
			IssuerURL: issuer,
			ClientID:  config.ClientID,
		},
		httpClient: &http.Client{Timeout: 10 * time.Second},
		jwksURL:    issuer + "/jwks/",
	}
}

// ValidateToken validates a JWT token and returns the claims
func (v *OIDCValidator) ValidateToken(ctx context.Context, tokenString string) (*OIDCClaims, error) {
	// Parse without verification first to get the key ID
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &OIDCClaims{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Get the key ID from header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: missing kid in token header", ErrInvalidToken)
	}

	// Get the public key
	publicKey, err := v.getPublicKey(ctx, kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse and validate the token
	claims := &OIDCClaims{}
	token, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate issuer (normalize trailing slash)
	expectedIssuer := strings.TrimSuffix(v.config.IssuerURL, "/")
	actualIssuer := strings.TrimSuffix(claims.Issuer, "/")
	if actualIssuer != expectedIssuer {
		return nil, fmt.Errorf("%w: expected %s, got %s", ErrInvalidIssuer, v.config.IssuerURL, claims.Issuer)
	}

	// Validate audience (client ID)
	if !v.validateAudience(claims.Audience) {
		return nil, fmt.Errorf("%w: expected %s", ErrInvalidAudience, v.config.ClientID)
	}

	return claims, nil
}

// validateAudience checks if the client ID is in the audience claim
func (v *OIDCValidator) validateAudience(audiences jwt.ClaimStrings) bool {
	for _, aud := range audiences {
		if aud == v.config.ClientID {
			return true
		}
	}
	return false
}

// getPublicKey retrieves the public key for the given key ID
func (v *OIDCValidator) getPublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	jwks, err := v.getJWKS(ctx)
	if err != nil {
		return nil, err
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid {
			return parseRSAPublicKey(key)
		}
	}

	// Key not found, try refreshing JWKS
	v.mu.Lock()
	v.jwksExpiry = time.Time{} // Force refresh
	v.mu.Unlock()

	jwks, err = v.getJWKS(ctx)
	if err != nil {
		return nil, err
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid {
			return parseRSAPublicKey(key)
		}
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

// getJWKS retrieves the JWKS, using cache if available
func (v *OIDCValidator) getJWKS(ctx context.Context) (*JWKSet, error) {
	v.mu.RLock()
	if v.jwks != nil && time.Now().Before(v.jwksExpiry) {
		jwks := v.jwks
		v.mu.RUnlock()
		return jwks, nil
	}
	v.mu.RUnlock()

	// Fetch fresh JWKS
	v.mu.Lock()
	defer v.mu.Unlock()

	// Double-check after acquiring write lock
	if v.jwks != nil && time.Now().Before(v.jwksExpiry) {
		return v.jwks, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("JWKS endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var jwks JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	v.jwks = &jwks
	v.jwksExpiry = time.Now().Add(1 * time.Hour) // Cache for 1 hour

	log.Printf("[OIDC] Refreshed JWKS from %s, found %d keys", v.jwksURL, len(jwks.Keys))

	return &jwks, nil
}

// parseRSAPublicKey converts a JWK to an RSA public key
func parseRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	if jwk.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", jwk.Kty)
	}

	// Decode N (modulus)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	// Decode E (exponent)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %w", err)
	}

	// Convert exponent bytes to int
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

// ExtractBearerToken extracts the token from Authorization header
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrNoAuthHeader
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrInvalidAuthHeader
	}

	return parts[1], nil
}
