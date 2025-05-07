package store

import "time"

// TokenRevoker defines the interface for token revocation operations
type TokenRevoker interface {
	// RevokeToken adds a token to the revocation list with a specified TTL
	RevokeToken(jti string, ttl time.Duration) error

	// IsRevoked checks if a token is in the revocation list
	IsRevoked(jti string) bool
}
