package s3

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
)

type RandIntFunc func(io.Reader, *big.Int) (*big.Int, error)

var DefaultRandIntFunc = rand.Int

const (
	numRandomCharsInAccessKey = 10
	numRandomCharsInSecretKey = 40
)

var (
	accessKeyLetters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	secretKeyLetters = []rune("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func GenerateAccessKey(randInt RandIntFunc) ([]byte, error) {
	accessKey, err := generateRandomKey(randInt, numRandomCharsInAccessKey, accessKeyLetters)
	if err != nil {
		return nil, err
	}

	return []byte("AKIA" + string(accessKey)), nil
}

func GenerateSecretKey(randInt RandIntFunc) ([]byte, error) {
	return generateRandomKey(randInt, numRandomCharsInSecretKey, secretKeyLetters)
}

func generateRandomKey(randInt RandIntFunc, numChars int, letters []rune) ([]byte, error) {
	key := make([]rune, numChars)

	for i := range key {
		num, err := randInt(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return nil, fmt.Errorf("failed to get random number: %w", err)
		}

		key[i] = letters[num.Int64()]
	}

	return []byte(string(key)), nil
}
