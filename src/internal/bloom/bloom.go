package bloom

import (
	"encoding/binary"
	"math"
)

const (
	// We use uint32 for bucket size internally
	bytesPerBucket = 4
)

func optimalSubhashes(numBuckets int, elementCount int) int {
	return int(math.Ceil(float64(numBuckets) / float64(elementCount) * math.Ln2))
}

func requiredBuckets(elementCount int, falsePositiveRate float64) int {
	return int(math.Ceil(-(float64(elementCount) * math.Log(falsePositiveRate) / (math.Pow(math.Ln2, 2)))))
}

// FilterSizeForFalsePositiveRate returns the memory footprint for a filter with
// the given constraints.
func FilterSizeForFalsePositiveRate(falsePositiveRate float64, elementCount int) int {
	return requiredBuckets(elementCount, falsePositiveRate) * bytesPerBucket
}

// NewFilterWithSize constructs a new bloom filter with the given constraints.
//  * 'bytes' - the number of bytes to allocate for the filter buckets
//  * 'elementCount' - the expected number of elements that will be present in
//      the filter
// Given these, the filter will choose the optimal number of hashes to perform
// to get the best possible false-positive rate.
func NewFilterWithSize(bytes int, elementCount int) *BloomFilter {
	if elementCount <= 0 {
		panic("Element count must be greater than 0")
	}
	if bytes <= 0 {
		panic("Bytes must be greater than 0")
	}

	numBuckets := bytes / bytesPerBucket
	filter := &BloomFilter{
		NumSubhashes: uint32(optimalSubhashes(numBuckets, elementCount)),
		Buckets:      make([]uint32, numBuckets),
	}

	return filter
}

// NewFilterWithFalsePositiveRate constructs a new bloom filter with the given
// constraints:
//  * 'falsePositiveRate' - the expected false positive rate when the filter is
//      loaded with 'elementCount' elements
//  * 'elementCount' - the expected number of elements that will be present in
//      the filter
//  * 'maxBytes' - the maximum memory footprint to allocate for this filter.
// Given these, the filter will choose a size that will provide the specified
// constraints.  If the memory footprint is capped by 'maxBytes', the
// falsePositiveRate will increase.
func NewFilterWithFalsePositiveRate(falsePositiveRate float64, elementCount int, maxBytes int) *BloomFilter {
	if falsePositiveRate <= 0.0 || falsePositiveRate >= 1.0 {
		panic("Expected false positive rate to be a value between 0 and 1")
	}
	if elementCount <= 0 {
		panic("Element count must be greater than 0")
	}
	if maxBytes <= 0 {
		panic("Max bytes must be greater than 0")
	}

	maxBuckets := maxBytes / bytesPerBucket
	numBuckets := int(math.Min(float64(requiredBuckets(elementCount, falsePositiveRate)), float64(maxBuckets)))

	filter := &BloomFilter{
		NumSubhashes: uint32(optimalSubhashes(numBuckets, elementCount)),
		Buckets:      make([]uint32, numBuckets),
	}

	return filter
}

// FalsePositiveRate returns the expected rate of false positives if the filter
// contains the given number of elements.
func (f *BloomFilter) FalsePositiveRate(elementCount int) float64 {
	// m: number of buckets, k: number of hashes
	m := float64(len(f.Buckets))
	k := float64(f.NumSubhashes)
	return math.Pow(1-math.Pow(1-1/m, k*float64(elementCount)), k)
}

// OverflowRate iterates over all buckets in the filter and returns the ratio
// of buckets that are in overflow to the total number of buckets.  Once a
// bucket overflows, it will never return to normal, even if all items
// corresponding to that bucket are removed - a new filter must be constructed
// instead.
func (f *BloomFilter) OverflowRate() float64 {
	numOverflow := 0
	for _, value := range f.Buckets {
		if value == math.MaxUint32 {
			numOverflow++
		}
	}
	return float64(numOverflow) / float64(len(f.Buckets))
}

// Add adds an item to the filter.
func (f *BloomFilter) Add(hash []byte) {
	f.forEachSubhash(hash, func(i int) {
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
	f.forEachSubhash(hash, func(i int) {
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
func (f *BloomFilter) UpperBoundCount(hash []byte) int {
	lowerBound := int(math.MaxUint32)
	f.forEachSubhash(hash, func(i int) {
		value := int(f.Buckets[i])
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

func (f *BloomFilter) forEachSubhash(hash []byte, cb func(int)) {
	// This is required so that we can guarantee hash generation in overflow
	if len(hash) < 8 {
		panic("Bloom filter hashes must be at least 8 bytes")
	}

	index := uint(0)

	overflowInitialized := false
	var overflowFirst, overflowSecond uint32

	for hashes := uint32(0); hashes < f.NumSubhashes; hashes++ {
		var value uint32
		if index+4 < uint(len(hash)) {
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

		cb(int(value) % len(f.Buckets))
	}
}
