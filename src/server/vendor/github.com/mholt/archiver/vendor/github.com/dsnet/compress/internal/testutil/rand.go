// Copyright 2015, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package testutil

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
)

// Rand implements a deterministic pseudo-random number generator.
// This differs from the math.Rand in that the exact output will be consistent
// across different versions of Go.
type Rand struct {
	cipher.Block
	blk [aes.BlockSize]byte
}

func NewRand(seed int) *Rand {
	var key [aes.BlockSize]byte
	binary.LittleEndian.PutUint64(key[:], uint64(seed))
	r, _ := aes.NewCipher(key[:])
	return &Rand{Block: r}
}

func (r *Rand) Int() int {
	r.Encrypt(r.blk[:], r.blk[:])
	return int(binary.BigEndian.Uint32(r.blk[:]) >> 1)
}

func (r *Rand) Intn(n int) int {
	return r.Int() % n
}

func (r *Rand) Bytes(n int) []byte {
	b := make([]byte, n)
	bb := b
	for len(bb) > 0 {
		r.Encrypt(r.blk[:], r.blk[:])
		cnt := copy(bb, r.blk[:])
		bb = bb[cnt:]
	}
	return b
}

func (r *Rand) Perm(n int) []int {
	m := make([]int, n)
	for i := 0; i < n; i++ {
		j := r.Intn(i + 1)
		m[i] = m[j]
		m[j] = i
	}
	return m
}
