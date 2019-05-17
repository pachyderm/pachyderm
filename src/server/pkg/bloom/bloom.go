package bloom

import (
	"encoding/binary"
	"fmt"
	"math"
)

func (f *BloomFilter) ExpectedFalsePositiveRate(numItems uint32) float32 {
	return math.Pow(math.E, float64(len(f.Buckets))/(-float64(numItems))*math.Pow(math.Ln2, 2))
}

func NewBloomFilter(hashLength uint32, numSubhashes uint32, numBuckets uint32) (*BloomFilter, error) {
	filter := &BloomFilter{
		HashLength:      hashLength,
		BytesPerSubhash: hashLength / numSubhashes,
		Buckets:         make([]uint32, numBuckets),
	}

	if err := filter.Validate(); err != nil {
		return nil, err
	}
	return filter, nil
}

func (f *BloomFilter) Validate() error {
	optimalHashes := math.Log2(float64(len(f.Buckets)))
	fmt.Printf("optimalHashes: %f\n", optimalHashes)
	fmt.Printf("bytesPerSubhash: %d\n", f.BytesPerSubhash)

	// require numBuckets to be a power of two so we don't have weighted areas of the hash space
	if len(f.Buckets)&(len(f.Buckets)-1) != 0 {
		return fmt.Errorf("Bloom filter number of buckets must be a power of two")
	}
	if math.Floor(float64(f.HashLength)/float64(f.BytesPerSubhash)) < optimalHashes {
		return fmt.Errorf("Bloom filter hash length is too short for the number of buckets")
	}
	return nil
}

// OverflowRate iterates over all buckets in the filter and returns the ratio
// of buckets that are in overflow to the total number of buckets.  Once a
// bucket overflows, it will never return to normal, even if all items
// corresponding to that bucket are removed - a new filter must be constructed
// instead.
func (f *BloomFilter) OverflowRate() float32 {
	numOverflow := 0
	for _, value := range f.Buckets {
		if value == math.MaxUint32 {
			numOverflow += 1
		}
	}
	return float32(numOverflow) / float32(len(f.Buckets))
}

// Add adds an item to the filter.
func (f *BloomFilter) Add(hash []byte) {
	f.forEachSubhash(hash, func(i uint32) {
		// If we overflow the bucket, we don't want it to wrap back to zero, just
		// leave it at MaxUint32 and it can never be removed from the filter.
		value := f.Buckets[i]
		if value != math.MaxUint32 {
			f.Buckets[i] = value + 1
		}
	})
}

// Remove removes an item from the filter
func (f *BloomFilter) Remove(hash []byte) {
	f.forEachSubhash(hash, func(i uint32) {
		// We can't remove from an overflowed bucket, as the information has already
		// been lost - just leave it at MaxUint32
		value := f.Buckets[i]
		if value != 0 && value != math.MaxUint32 {
			f.Buckets[i] = value - 1
		}
	})
}

// UpperBoundCount returns the maximum number of items in the filter matching
// the given hash.
func (f *BloomFilter) UpperBoundCount(hash []byte) uint32 {
	lowerBound := uint32(math.MaxUint32)
	f.forEachSubhash(hash, func(i uint32) {
		value := f.Buckets[i]
		if value < lowerBound {
			lowerBound = value
		}
	})
	return lowerBound
}

// IsNotPresent returns true if an item corresponding to the given hash has
// definitely not been added to the set.
func (f *BloomFilter) IsNotPresent(hash []byte) bool {
	return f.UpperBoundCount(hash) == 0
}

func (f *BloomFilter) forEachSubhash(hash []byte, cb func(uint32)) {
	if uint32(len(hash)) != f.HashLength {
		panic(fmt.Sprintf("Incorrect hash length passed to bloom filter, expected %d, got %d", f.HashLength, len(hash)))
	}

	fmt.Printf("Hash: %s\n", hash)
	// Divide the hash into equal chunks based on the number of buckets
	for i := uint32(0); i <= f.HashLength-f.BytesPerSubhash; i += f.BytesPerSubhash {
		subhash := hash[i : i+f.BytesPerSubhash]
		fmt.Printf("Hash[%d:%d]: %s\n", i, i+f.BytesPerSubhash, subhash)
		value := binary.LittleEndian.Uint32(subhash)
		fmt.Printf("Vaule: %d\n", value)
		cb(value % uint32(len(f.Buckets)))
	}
}
