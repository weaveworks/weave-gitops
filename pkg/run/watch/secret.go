package watch

import (
	"math/rand"
)

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func generateAccessKey(numChars int, seed int64) []byte {
	rand.Seed(seed)

	secretKey := make([]byte, numChars)
	for i := range secretKey {
		secretKey[i] = letters[rand.Intn(len(letters))]
	}

	return secretKey
}

func generateSecretKey(numChars int, seed int64) []byte {
	rand.Seed(seed)

	secretKey := make([]byte, numChars)
	for i := range secretKey {
		secretKey[i] = letters[rand.Intn(len(letters))]
	}

	return secretKey
}
