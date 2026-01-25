package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
	ErrIssuer       = errors.New("invalid issuer")
	ErrAudience     = errors.New("invalid audience")
)

// Claims carries JWT claim fields used by the platform.
type Claims struct {
	TenantID  string   `json:"tenant_id"`
	Subject   string   `json:"sub,omitempty"`
	Issuer    string   `json:"iss,omitempty"`
	Audience  any      `json:"aud,omitempty"`
	Expiry    int64    `json:"exp,omitempty"`
	NotBefore int64    `json:"nbf,omitempty"`
	IssuedAt  int64    `json:"iat,omitempty"`
	Roles     []string `json:"roles,omitempty"`
}

type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// ParseHS256 verifies a HS256 JWT and returns claims.
func ParseHS256(token string, secret string, issuer string, audience string, now time.Time) (*Claims, error) {
	if secret == "" {
		return nil, ErrInvalidToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var hdr header
	if err := json.Unmarshal(headerBytes, &hdr); err != nil {
		return nil, ErrInvalidToken
	}
	if hdr.Alg != "HS256" {
		return nil, ErrInvalidToken
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if !verifySignature(secret, parts[0], parts[1], parts[2]) {
		return nil, ErrInvalidToken
	}
	if claims.Expiry != 0 && now.Unix() >= claims.Expiry {
		return nil, ErrExpiredToken
	}
	if claims.NotBefore != 0 && now.Unix() < claims.NotBefore {
		return nil, ErrInvalidToken
	}
	if issuer != "" && claims.Issuer != issuer {
		return nil, ErrIssuer
	}
	if audience != "" && !audienceMatch(claims.Audience, audience) {
		return nil, ErrAudience
	}
	return &claims, nil
}

func verifySignature(secret string, header string, payload string, signature string) bool {
	signatureBytes, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(header))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write([]byte(payload))
	expected := mac.Sum(nil)
	return hmac.Equal(signatureBytes, expected)
}

// SignHS256 signs claims into a HS256 JWT string.
func SignHS256(claims Claims, secret string) (string, error) {
	if secret == "" {
		return "", ErrInvalidToken
	}
	headerBytes, err := json.Marshal(header{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", ErrInvalidToken
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", ErrInvalidToken
	}
	headerEnc := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadEnc := base64.RawURLEncoding.EncodeToString(payloadBytes)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(headerEnc))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write([]byte(payloadEnc))
	signature := mac.Sum(nil)
	signatureEnc := base64.RawURLEncoding.EncodeToString(signature)
	return headerEnc + "." + payloadEnc + "." + signatureEnc, nil
}

func audienceMatch(aud any, expected string) bool {
	switch value := aud.(type) {
	case string:
		return value == expected
	case []any:
		for _, item := range value {
			if str, ok := item.(string); ok && str == expected {
				return true
			}
		}
		return false
	case []string:
		for _, item := range value {
			if item == expected {
				return true
			}
		}
		return false
	default:
		return false
	}
}
