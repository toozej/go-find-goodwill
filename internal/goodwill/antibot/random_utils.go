package antibot

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// getRandomInt generates a random integer between 0 and max-1
func getRandomInt(max int) (int, error) {
	if max <= 0 {
		return 0, nil
	}

	// Generate random int64 using crypto/rand
	maxInt := big.NewInt(int64(max))
	randomInt, err := rand.Int(rand.Reader, maxInt)
	if err != nil {
		return 0, fmt.Errorf("failed to generate random number: %w", err)
	}

	return int(randomInt.Int64()), nil
}

// getRandomFloat64 generates a random float64 between 0 and 1 using crypto/rand
func getRandomFloat64() (float64, error) {
	// Generate a random 64-bit integer
	randomInt, err := rand.Int(rand.Reader, big.NewInt(1<<53)) // 53 bits for IEEE 754 double precision
	if err != nil {
		return 0.0, fmt.Errorf("failed to generate random number: %w", err)
	}

	// Convert to float64 in range [0, 1)
	return float64(randomInt.Int64()) / float64(1<<53), nil
}
