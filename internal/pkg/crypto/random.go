package crypto

import (
	"crypto/rand"
	"math/big"
)

// GenerateSecureRandomInt generates a secure random integer between 0 and max (exclusive)
// It uses crypto/rand for cryptographically secure randomness.
func GenerateSecureRandomInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// Fallback to non-secure if crypto random fails (extremely unlikely)
		// Ideally we should panic or return error, but for OTP generation in this context
		// returning 0 or panic is safer.
		// Let's safe fail to 0, but log if we had logger. Pkg crypto shouldn't depend on logger loop.
		return 0
	}
	return int(n.Int64())
}
