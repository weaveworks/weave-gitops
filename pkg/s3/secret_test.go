package s3

import (
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"testing"

	. "github.com/onsi/gomega"
)

func deterministicRandInt(seed int64, err error) RandIntFunc {
	seeded := false

	return func(_ io.Reader, max *big.Int) (*big.Int, error) {
		if err != nil {
			return nil, err
		}

		if !seeded {
			rand.Seed(seed)

			seeded = true
		}

		return big.NewInt(int64(rand.Intn(int(max.Int64())))), nil
	}
}

func TestGenerators(t *testing.T) {
	tests := []struct {
		name        string
		generator   func(RandIntFunc) ([]byte, error)
		randIntFunc RandIntFunc
		expected    string
		expectedErr bool
	}{
		{
			name:        "GenerateAccessKey generates a deterministic access key",
			generator:   GenerateAccessKey,
			randIntFunc: deterministicRandInt(100, nil),
			expected:    "AKIA5UQA4UZJM3",
			expectedErr: false,
		},
		{
			name:        "GenerateAccessKey properly returns an error if RNG fails",
			generator:   GenerateAccessKey,
			randIntFunc: deterministicRandInt(0, fmt.Errorf("foobar")),
			expected:    "",
			expectedErr: true,
		},
		{
			name:        "GenerateSecretKey generates a deterministic secret key",
			generator:   GenerateSecretKey,
			randIntFunc: deterministicRandInt(512, nil),
			expected:    "Fg5n9W6CwTfnMu4FzEk8xuTomwk2OpFe0yLcLMAL",
			expectedErr: false,
		},
		{
			name:        "GenerateSecretKey properly returns an error if RNG fails",
			generator:   GenerateSecretKey,
			randIntFunc: deterministicRandInt(0, fmt.Errorf("foobar")),
			expected:    "",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			accessKey, err := tt.generator(tt.randIntFunc)
			if tt.expectedErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
			g.Expect(string(accessKey)).To(Equal(tt.expected))
		})
	}
}
