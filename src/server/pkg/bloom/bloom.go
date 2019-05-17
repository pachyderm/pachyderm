package bloom

import (
	"math"
)

type Filter struct {
	data *BloomData
}

func NewFilter(uint32 hashLength, uint32 numHashes, uint32 numBuckets) Filter {
	return &Filter{
		data: &BloomData{
			HashLength: hashLength,
			NumHashes: numHashes,
			Buckets: make([]uint32, numBuckets),
		}
	}
}

func FromData(data *BloomData) {
	return &Filter{data: data}
}

func (f *Filter) Data() *BloomData {
	return f.data
}

func (f *Filter) Add(hash []byte) {
	f.forEachHash(hash, func(i uint32) {
		// If we overflow the bucket, we don't want it to wrap back to zero, just
		// leave it at MaxUint32 and it can never be removed from the filter.
		val := f.data.buckets[i]
		if val != math.MaxUint32 {
			f.data.buckets[i] = val + 1
		}
	})
}

func (f *Filter) Remove(hash []byte) {
	f.forEachHash(hash, func(i uint32) {
		// We can't remove from an overflowed bucket, as the information has already
		// been lost - just leave it at MaxUint32
		val := f.data.buckets[i]
		if val == 0 {
			panic("Removing an item from a bloom filter that was not inserted")
		} else if val != math.MaxUint32 {
			f.data.buckets[i] = val - 1
		}
	})
}

func (f *Filter) UpperBoundCount(hash []byte) {
	lowerBound := math.MaxUint32
	f.forEachHash(hash, func(i uint32) {
		val := f.data.buckets[i]
		if val < lowerBound {
			lowerBound = val
		}
	})
}

func (f *Filter) IsNotPresent(hash []byte) {
	return f.UpperBoundCount(hash) == 0
}

func (f *Filter) forEachHash(hash []byte, f func(uint32)) {
	if len(hash) != f.data.HashLength {
		panic(fmt.Sprintf("Incorrect hash length passed to bloom filter, expected %d, got %d", f.data.HashLength, len(hash)))
	}
	// Divide the hash into equal chunks based on the number of buckets
}
