package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// nonces stores issued nonces with expiry (5 min). Kept in-memory — fine for a single instance.
var (
	nonceMu sync.Mutex
	nonces  = map[string]time.Time{}
)

// Nonce issues a one-time random nonce.
func Nonce() http.HandlerFunc {
	// Prune expired nonces every minute
	go func() {
		for range time.Tick(time.Minute) {
			nonceMu.Lock()
			for n, exp := range nonces {
				if time.Now().After(exp) {
					delete(nonces, n)
				}
			}
			nonceMu.Unlock()
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			http.Error(w, "failed to generate nonce", http.StatusInternalServerError)
			return
		}
		nonce := hex.EncodeToString(b)

		nonceMu.Lock()
		nonces[nonce] = time.Now().Add(5 * time.Minute)
		nonceMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"nonce": nonce})
	}
}

// consumeNonce returns true and deletes the nonce if it exists and hasn't expired.
func consumeNonce(nonce string) bool {
	nonceMu.Lock()
	defer nonceMu.Unlock()
	exp, ok := nonces[nonce]
	if !ok || time.Now().After(exp) {
		return false
	}
	delete(nonces, nonce)
	return true
}
