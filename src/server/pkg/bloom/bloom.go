package bloom

import (
	"encoding/binary"
	"math"
)

func optimalSubhashes(numBuckets uint32, expectedCount uint32) uint32 {
	return uint32(math.Ceil(float64(numBuckets) / float64(expectedCount) * math.Ln2))
}

func requiredBuckets(elementCount uint32, falsePositiveRate float64) uint32 {
	return uint32(math.Ceil(-(float64(elementCount) * math.Log(falsePositiveRate) / (math.Pow(math.Ln2, 2)))))
}

// FilterSize returns the memory footprint for a filter with the given constraints
func FilterSizeForFalsePositiveRate(falsePositiveRate float64, elementCount uint32) uint32 {
	return requiredBuckets(elementCount, falsePositiveRate) * 4
}

// NewFilterWithSize constructs a new bloom filter with the given constraints.
//  * 'bytes' - the number of bytes to allocate for the filter buckets
//  * 'elementCount' - the expected number of elements that will be present in
//      the filter
// Given these, the filter will choose the optimal number of hashes to perform
// to get the best possible false-positive rate.
func NewFilterWithSize(bytes uint32, elementCount uint32) *BloomFilter {
	numBuckets := bytes

	filter := &BloomFilter{
		NumSubhashes: optimalSubhashes(numBuckets, elementCount),
		Buckets:      make([]byte, numBuckets),
	}

	return filter
}

// NewFilterWithFalsePositiveRate constructs a new bloom filter with the given
// constraints:
//  * 'falsePositiveRate' - the expected false positive rate when the filter is
//      loaded with 'elementCount' elements
//  * 'elementCount' - the expected number of elements that will be present in
//      the filter
// Given these, the filter will choose a size that will provide the specified
// constraints.  As a sanity check, `FilterSizeForFalsePositiveRate` should be
// called beforehand to avoid an accidental ridiculously large allocation.
func NewFilterWithFalsePositiveRate(falsePositiveRate float64, elementCount uint32) *BloomFilter {
	if falsePositiveRate < 0.0 || falsePositiveRate > 1.0 {
		panic("Expected false positive rate to be a value between 0 and 1")
	}

	numBuckets := requiredBuckets(elementCount, falsePositiveRate)

	filter := &BloomFilter{
		NumSubhashes: optimalSubhashes(numBuckets, elementCount),
		Buckets:      make([]byte, numBuckets),
	}

	return filter
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
		if value == math.MaxUint8 {
			numOverflow += 1
		}
	}
	return float32(numOverflow) / float32(len(f.Buckets))
}

// Add adds an item to the filter.
func (f *BloomFilter) Add(hash []byte) {
	f.forEachSubhash(hash, func(i uint32) {
		// If we overflow the bucket, we don't want it to wrap back to zero, just
		// leave it at MaxUint8 and it can never be removed from the filter.
		value := f.Buckets[i]
		if value != math.MaxUint8 {
			f.Buckets[i] = value + 1
		}
	})
}

// Remove removes an item from the filter
func (f *BloomFilter) Remove(hash []byte) {
	f.forEachSubhash(hash, func(i uint32) {
		// We can't remove from an overflowed bucket, as the information has already
		// been lost - just leave it at MaxUint8
		value := f.Buckets[i]
		if value != 0 && value != math.MaxUint8 {
			f.Buckets[i] = value - 1
		}
	})
}

// UpperBoundCount returns the maximum number of items in the filter matching
// the given hash.
func (f *BloomFilter) UpperBoundCount(hash []byte) uint8 {
	lowerBound := uint8(math.MaxUint8)
	f.forEachSubhash(hash, func(i uint32) {
		value := uint8(f.Buckets[i])
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
	// This is required so that we can guarantee hash generation in overflow
	if len(hash) < 8 {
		panic("Bloom filter hashes must be at least 8 bytes")
	}

	index := uint32(0)

	overflowInitialized := false
	var overflowFirst, overflowSecond uint32

	for hashes := uint32(0); hashes < f.NumSubhashes; hashes += 1 {
		var value uint32
		if index+4 < uint32(len(hash)) {
			subhash := hash[index : index+4]
			value = binary.LittleEndian.Uint32(subhash)
			index += 4
		} else {
			// This should be an ok way to extend the hash, via https://www.eecs.harvard.edu/~michaelm/postscripts/tr-02-05.pdf
			if !overflowInitialized {
				subhash := hash[0:4]
				overflowFirst = binary.LittleEndian.Uint32(subhash)
				subhash = hash[4:8]
				overflowSecond = binary.LittleEndian.Uint32(subhash)
			}
			value = overflowFirst + hashes*(overflowSecond)
		}

		cb(value % uint32(len(f.Buckets)))
	}
}
