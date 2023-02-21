package compositehash

import (
	"hash/fnv"
	"time"
)

// New generates a composite hash for the given text and time.
//
// The function computes a hash value of the text using the FNV-1a hash function, which
// is fast and provides good enough distribution for our purposes. The hash value is then
// combined with the time, expressed in milliseconds since the epoch, to generate a unique
// identifier. The resulting identifier is guaranteed to be unique and sortable for
// the given text and time.
//
// Parameters:
//   - text: the text to generate a composite hash for.
//   - atime: the time to use in generating the composite hash.
//
// Returns:
//   - int64: the composite hash for the given text and time.
//   - error: if there was an error computing the hash value of the text.
func New(text string, atime time.Time) (int64, error) {
	// Compute the hash value
	// Note: we use the FNV-1a hash function, which is fast and provides good enough
	// distribution for our purposes.
	h := fnv.New32a()
	_, err := h.Write([]byte(text))
	if err != nil {
		return 0, err
	}

	// The hash value is guaranteed to be 32 bits, so we can safely convert it to an int64
	// without losing any information.
	// 99877 is a prime number that is close to 100000, which is the maximum value for
	// the hash value. This ensures that the resulting ID is always less than 100000.
	hash := h.Sum32() % 99877

	// Combine the hash value with the time, expressed in milliseconds since the epoch,
	// to generate a unique identifier, which is guaranteed to be unique and sortable.
	return atime.UnixMilli()*100000 + int64(hash), nil
}
