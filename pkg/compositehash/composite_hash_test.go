package compositehash

import (
	"hash/fnv"
	"math/big"
	"testing"
	"time"

	"github.com/onsi/gomega"
)

func TestNewCompositeHash(t *testing.T) {
	type testObj struct {
		Name  string
		Atime time.Time
	}

	// Use a fixed time value for the test
	obj := testObj{Name: "Alice", Atime: time.Date(2023, time.February, 21, 10, 30, 0, 0, time.UTC)}

	// Use gomega for assertions
	g := gomega.NewGomegaWithT(t)

	// Compute the hash value
	id, err := New(obj.Name, obj.Atime)

	// Check that the hash value is as expected
	h := fnv.New32a()
	_, _ = h.Write([]byte(obj.Name))
	expectedContentHash := int64(h.Sum32())
	hash := new(big.Int).Mod(big.NewInt(expectedContentHash), big.NewInt(99877))
	expectedID := obj.Atime.UnixMilli()*100000 + hash.Int64()

	// Ensure the function didn't return an error
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Ensure the ID is as expected
	g.Expect(id).To(gomega.Equal(expectedID))
}
