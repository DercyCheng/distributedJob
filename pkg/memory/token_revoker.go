package memory

import (
	"sync"
	"time"

	"distributedJob/internal/store"
)

// MemoryTokenRevoker implements TokenRevoker interface using memory map
type MemoryTokenRevoker struct {
	revokedTokens map[string]time.Time
	mutex         sync.RWMutex
}

// NewMemoryTokenRevoker creates a new memory token revoker
func NewMemoryTokenRevoker() store.TokenRevoker {
	revoker := &MemoryTokenRevoker{
		revokedTokens: make(map[string]time.Time),
	}

	// Start a cleanup goroutine to remove expired tokens
	go revoker.cleanupRoutine()

	return revoker
}

// RevokeToken adds a token to the revocation list with a specified TTL
func (m *MemoryTokenRevoker) RevokeToken(jti string, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Set the token expiration time
	expireTime := time.Now().Add(ttl)
	m.revokedTokens[jti] = expireTime

	return nil
}

// IsRevoked checks if a token is in the revocation list
func (m *MemoryTokenRevoker) IsRevoked(jti string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	expirationTime, exists := m.revokedTokens[jti]

	// Token is revoked if it exists and has not expired
	return exists && expirationTime.After(time.Now())
}

// cleanupRoutine periodically removes expired tokens
func (m *MemoryTokenRevoker) cleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		m.removeExpiredTokens()
	}
}

// removeExpiredTokens removes tokens that have passed their expiration time
func (m *MemoryTokenRevoker) removeExpiredTokens() {
	now := time.Now()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for jti, expirationTime := range m.revokedTokens {
		if now.After(expirationTime) {
			delete(m.revokedTokens, jti)
		}
	}
}
