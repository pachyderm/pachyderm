package bloom

import (
	"encoding/binary"
	"fmt"
	"math"
)

func optimalSubhashes(numBuckets uint32, expectedCount uint32) uint32 {
	return uint32(math.Ceil(float64(numBuckets) / float64(expectedCount) * math.Ln2))
}

func requiredBuckets(elementCount uint32, falsePositiveRate float32) uint32 {
	return uint32(math.Ceil(-(float64(elementCount) * math.Ln(falsePositiveRate) / (math.Pow(math.Ln2, 2)))))
}

func bytesPerSubhash(numBuckets uint32) uint32 {
	return uint32(math.Ceil(math.Ln(numBuckets) / math.Ln(2)))
}

// FilterSize returns the memory footprint for a filter with the given constraints
func FilterSizeForFalsePositiveRate(falsePositiveRate float32, elementCount uint32) uint32 {
	return requiredBuckets(elementCount, falsePositiveRate) * 4
}

// NewFilterWithSize constructs a new bloom filter with the given constraints.
//  * 'bytes' - the number of bytes to allocate for the filter buckets
//  * 'elementCount' - the expected number of elements that will be present in
//      the filter
// Given these, the filter will choose the optimal number of hashes to perform
// to get the best possible false-positive rate.
func NewFilterWithSize(bytes uint32, elementCount uint32, hashLength uint32) *BloomFilter {
	numBuckets := constraints.Bytes / 4
	subhashLength := bytesPerSubhash(numBuckets)

	filter := &BloomFilter{
		NumSubhashes:    optimalSubhashes(numBuckets, elementCount),
		BytesPerSubhash: subhashLength,
		Buckets:         make([]uint32, numBuckets),
	}

	return filter
}

func NewFilterWithFalsePositiveRate(falsePositiveRate float32, elementCount uint32, hashLength uint32) *BloomFilter {
	if falsePositiveRate < 0.0 || falsePositiveRate > 1.0 {
		panic("Expected false positive rate to be a value between 0 and 1")
	}

	numBuckets := requiredBuckets(elementCount, falsePositiveRate)
	subhashLength := bytesPerSubhash(numBuckets)

	filter := &BloomFilter{
		NumSubhashes:    optimalSubhashes(numBuckets, elementCount),
		BytesPerSubhash: subhashLength,
		Buckets:         make([]uint32, numBuckets),
	}

	return filter, nil
}

func (f *BloomFilter) FalsePositiveRate(numItems uint32) float64 {
	// m: number of buckets, k: number of hashes
	m := float64(len(f.Buckets))
	k := float64(f.NumSubhashes)
	return math.Pow(1-math.Pow(1-1/m, k*float64(numItems)), k)
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
	if len(hash) < 8 {
		panic("Bloom filter hashes must be at least 8 bytes")
	}

	index := 0
	var first, second *uint32
	for hashes := uint32(0); hashes < f.NumSubhashes; hashes += 1 {
		var value uint32
		if index+f.BytesPerSubhash < len(hash) {
			subhash := hash[i : i+f.BytesPerSubhash]
			value = binary.LittleEndian.Uint32(subhash)
		} else {
			// This should be an ok way to extend the hash, via https://www.eecs.harvard.edu/~michaelm/postscripts/tr-02-05.pdf
			if first == nil {
				subhash := hash[0:4]
				first = &binary.LittleEndian.Uint32(subhash)
				subhash = hash[4:8]
				second = &binary.LittleEndian.Uint32(subhash)
			}
			value = *first + hashes*(*second)
		}

		fmt.Printf("Bucket: %d (%d)\n", value%uint32(len(f.Buckets)), value)
		cb(value % uint32(len(f.Buckets)))
	}
}
