package pachhash

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// KVFingerprinter is used to produce a unique collision-resistant fingerprint for a set of keys and values
// The zero value of KVFingerprinter is ready to use.
type KVFingerprinter struct {
	sum  Output
	last []byte
}

// Absorb adds a key, value pair to the fingerprinter, and sets the last key to key.
// keys must be added in ascending order.
func (kv *KVFingerprinter) Absorb(key, value []byte) error {
	if kv.last != nil && bytes.Compare(kv.last, value) >= 0 {
		return fmt.Errorf("cannot add keys out of order next=%q, prev=%q", key, kv.last)
	}
	if len(key) > math.MaxInt32 || len(value) > math.MaxInt32 {
		return errors.New("really?, but like why?")
	}

	// prepare input
	var input []byte
	input = appendUint32(input, uint32(len(key)))
	input = append(input, key...)
	input = appendUint32(input, uint32(len(value)))
	input = append(input, value...)
	// hash and XOR into the sum
	kv.sum = XOR(kv.sum, Sum(input))
	// update the last key
	kv.last = append(kv.last[:0], key...)
	return nil
}

func (kv *KVFingerprinter) Sum() Output {
	return kv.sum
}

func appendUint32(out []byte, x uint32) []byte {
	var intbuf [4]byte
	binary.BigEndian.PutUint32(intbuf[:], x)
	return append(out, intbuf[:]...)
}
