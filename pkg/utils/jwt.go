package utils

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func SignToken(userId, username, role string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET environment variable is not set")
	}

	jwtExpiresIn := os.Getenv("JWT_EXPIRES_IN")

	claims := jwt.MapClaims{
		"uid":  userId,
		"user": username,
		"role": role,
	}

	if jwtExpiresIn != "" {
		duration, err := time.ParseDuration(jwtExpiresIn)
		if err != nil {
			return "", errors.New("internal error")
		}
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(duration))
	} else {
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(15 * time.Minute))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// Acts as an in memory database where we store valid tokens
// Important to make it concurrency-safe since multiple requests may arrive at the same time
type JWTStore struct {
	mu     sync.Mutex
	Tokens map[string]time.Time
}

func (store *JWTStore) AddToken(token string, expiryTime time.Time) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.Tokens[token] = expiryTime
}

func (store *JWTStore) CleanUpExpiredTokens() {
	for {
		time.Sleep(2 * time.Minute)

		store.mu.Lock()
		for token, timestamp := range store.Tokens {
			if time.Now().After(timestamp) {
				delete(store.Tokens, token)
			}
		}
		store.mu.Unlock()
	}
}

// IsBlacklisted checks if a token has been blacklisted (logged out)
// Returns true if the token is in the blacklist (i.e., user has logged out)
func (store *JWTStore) IsBlacklisted(token string) bool {
	store.mu.Lock()
	defer store.mu.Unlock()

	_, ok := store.Tokens[token]
	return ok
}

var JwtStore = JWTStore{
	Tokens: make(map[string]time.Time),
}
