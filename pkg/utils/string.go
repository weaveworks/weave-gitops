package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	mrand "math/rand"
	"time"
)

const (
	alphaNumeric = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// GenerateRandomString will create a new random string with alphanumeric characters.
// The length can also vary by using the min and max parameters. To have a consistent length
// such as 11, you would pass (11, 12) for the min and max respectively
func GenerateRandomString(min, max int) (string, error) {
	mrand.Seed(time.Now().UnixNano())

	length := randInt(min, max)
	value := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphaNumeric))))
		if err != nil {
			return "", fmt.Errorf("error generated random string: %w", err)
		}

		value[i] = alphaNumeric[num.Int64()]
	}

	return string(value), nil
}

func randInt(min int, max int) int {
	return min + mrand.Intn(max-min)
}
