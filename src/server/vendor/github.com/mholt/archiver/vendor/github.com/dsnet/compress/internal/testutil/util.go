// Copyright 2015, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

// Package testutil is a collection of testing helper methods.
package testutil

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

// ResizeData resizes the input. If n < 0, then the original input will be
// returned as is. If n <= len(input), then the input slice will be truncated.
// However, if n > len(input), then the input will be replicated to fill in
// the missing bytes, but each replicated string will be XORed by some byte
// mask to avoid favoring algorithms with large LZ77 windows.
//
// If n > len(input), then len(input) must be > 0.
func ResizeData(input []byte, n int) []byte {
	if n < 0 {
		return input
	}
	if len(input) >= n {
		return input[:n]
	}
	if len(input) == 0 {
		panic("unable to replicate an empty string")
	}

	var mask byte
	output := make([]byte, n)
	for i := range output {
		idx := i % len(input)
		output[i] = input[idx] ^ mask
		if idx == len(input)-1 {
			mask++
		}
	}
	return output
}

// MustLoadFile must load a file or else panics.
func MustLoadFile(file string) []byte {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return b
}

// MustDecodeHex must decode a hexadecimal string or else panics.
func MustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// MustDecodeBitGen must decode a BitGen formatted string or else panics.
func MustDecodeBitGen(s string) []byte {
	b, err := DecodeBitGen(s)
	if err != nil {
		panic(err)
	}
	return b
}

// BytesCompare compares inputs a and b and reports whether they are equal.
//
// If they are not equal, it returns two one-line strings that are
// representative of the differences between the two strings.
// The output will be quoted strings if it seems like the data is text,
// otherwise, it will use hexadecimal strings.
//
// Example usage:
//
//	if got, want, ok := testutil.BytesCompare(output, v.output); !ok {
//		t.Errorf("output mismatch:\ngot  %s\nwant %s", got, want)
//	}
//
func BytesCompare(a, b []byte) (sa, sb string, ok bool) {
	if ok = bytes.Equal(a, b); ok {
		return
	}

	commonPrefix := func(a, b []byte) int {
		if len(a) > len(b) {
			a, b = b, a
		}
		for i := range a {
			if a[i] != b[i] {
				return i
			}
		}
		return len(a)
	}

	formatter := func(a, b []byte, format string, trimHead, maxLen int) (sa, sb string) {
		trimHead -= maxLen / 2 // Always provide context of equal bytes
		if trimHead < 0 {
			trimHead = 0
		}
		if trimHead > (len(a) - maxLen) {
			trimHead = (len(a) - maxLen)
		}
		if trimHead > (len(b) - maxLen) {
			trimHead = (len(b) - maxLen)
		}

		var head, atail, btail string
		if trimHead > 0 {
			a = a[trimHead:]
			b = b[trimHead:]
			head = fmt.Sprintf("(%d bytes)...", trimHead)
		}
		if len(a) > maxLen {
			atail = fmt.Sprintf("...(%d bytes)", len(a)-maxLen)
			a = a[:maxLen]
		}
		if len(b) > maxLen {
			btail = fmt.Sprintf("...(%d bytes)", len(b)-maxLen)
			b = b[:maxLen]
		}
		sa = fmt.Sprintf("%s"+format+"%s", head, a, atail)
		sb = fmt.Sprintf("%s"+format+"%s", head, b, btail)
		return sa, sb
	}

	const maxLen = 64
	n := commonPrefix(a, b)
	sa, sb = formatter(a, b, "%q", n, maxLen) // Favor quoted output, first
	if s := sa + sb; strings.Count(s, `\u`)+strings.Count(s, `\x`) > maxLen/8 {
		sa, sb = formatter(a, b, "%x", n, maxLen/2) // Fallback to hex, next
	}
	return sa, sb, false
}

// BuggyReader returns Err after N bytes have been read from R.
type BuggyReader struct {
	R   io.Reader
	N   int64 // Number of valid bytes to read
	Err error // Return this error after N bytes
}

func (br *BuggyReader) Read(buf []byte) (int, error) {
	if int64(len(buf)) > br.N {
		buf = buf[:br.N]
	}
	n, err := br.R.Read(buf)
	br.N -= int64(n)
	if err == nil && br.N <= 0 {
		return n, br.Err
	}
	return n, err
}

// BuggyWriter returns Err after N bytes have been written to W.
type BuggyWriter struct {
	W   io.Writer
	N   int64 // Number of valid bytes to write
	Err error // Return this error after N bytes
}

func (bw *BuggyWriter) Write(buf []byte) (int, error) {
	if int64(len(buf)) > bw.N {
		buf = buf[:bw.N]
	}
	n, err := bw.W.Write(buf)
	bw.N -= int64(n)
	if err == nil && bw.N <= 0 {
		return n, bw.Err
	}
	return n, err
}
