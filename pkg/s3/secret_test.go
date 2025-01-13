package s3

import (
	"fmt"
	"io"
	"math/big"
	"math/rand/v2"
	"testing"

	. "github.com/onsi/gomega"
)

func deterministicRandInt(seed uint64, err error) RandIntFunc {
	var srand *rand.Rand

	return func(_ io.Reader, max *big.Int) (*big.Int, error) {
		if err != nil {
			return nil, err
		}

		if srand == nil {
			srand = rand.New(rand.NewPCG(seed, 0)) // #nosec G404
		}

		return big.NewInt(srand.Int64N(max.Int64())), nil // #nosec G404
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
			expected:    "AKIATBK3988IAG",
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
			expected:    "0aEEdyKByGEXsQUh1af86o6HON4Ig468I6DhJH1C",
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
