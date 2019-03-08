// Copyright 2016, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package testutil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/dsnet/compress/internal"
)

// DecodeBitGen decodes a BitGen formatted string.
//
// The BitGen format allows bit-streams to be generated from a series of tokens
// describing bits in the resulting string. The format is designed for testing
// purposes by aiding a human in the manual scripting of a compression stream
// from individual bit-strings. It is designed to be relatively succinct, but
// allow the user to have control over the bit-order and also to allow the
// presence of comments to encode authorial intent.
//
// The format consists of a series of tokens delimited by either whitespace,
// a group marker, or a comment. As a special case, delimiters within a quoted
// string do not separate tokens. The '#' character is used for commenting
// such that any remaining bytes on the given line is ignored.
//
// The first valid token must either be a "<<<" (little-endian) or a ">>>"
// (big-endian). This determines whether the preceding bits in the stream are
// packed starting with the least-significant bits of a byte (little-endian) or
// packed starting with the most-significant bits of a byte (big-endian).
// Formats like DEFLATE and Brotli use little-endian, while BZip2 uses a
// big-endian bit-packing mode. This token appears exactly once at the start.
//
// A token of the form "<" (little-endian) or ">" (big-endian) determines the
// current bit-parsing mode, which alters the way subsequent tokens are
// processed. The format defaults to using a little-endian bit-parsing mode.
//
// A token of the pattern "[01]{1,64}" forms a bit-string (e.g. 11010).
// If the current bit-parsing mode is little-endian, then the right-most bits of
// the bit-string are written first to the resulting bit-stream. Likewise, if
// the bit-parsing mode is big-endian, then the left-most bits of the bit-string
// are written first to the resulting bit-stream.
//
// A token of the pattern "D[0-9]+:[0-9]+" or "H[0-9]+:[0-9a-fA-F]{1,16}"
// represents either a decimal value or a hexadecimal value, respectively.
// This numeric value is converted to the unsigned binary representation and
// used as the bit-string to write. The first number indicates the bit-length
// of the bit-string and must be between 0 and 64 bits. The second number
// represents the numeric value. The bit-length must be long enough to contain
// the resulting binary value. If the current bit-parsing mode is little-endian,
// then the least-significant bits of this binary number are written first to
// the resulting bit-stream. Likewise, the opposite holds for big-endian mode.
//
// A token that is of the pattern "X:[0-9a-fA-F]+" represents literal bytes in
// hexadecimal format that should be written to the resulting bit-stream.
// This token is affected by neither the bit-packing nor the bit-parsing modes.
// However, it may only be used when the bit-stream is currently byte-aligned.
//
// A token that begins and ends with a '"' represents literal bytes in human
// readable form. This double-quoted string is parsed in the same way that the
// Go language parses strings. Any delimiter tokens within the context of a
// quoted string have no effect. This token is affected by neither the
// bit-packing nor the bit-parsing modes. However, it may only be used when the
// bit-stream is already byte-aligned.
//
// A token decorator of "<" (little-endian) or ">" (big-endian) may begin
// any binary token or decimal token. This will affect the bit-parsing mode
// for that token only. It will not set the overall global mode (that still
// needs to be done by standalone "<" and ">" tokens). This decorator may not
// applied to the literal bytes token.
//
// A token decorator of the pattern "[*][0-9]+" may trail any token except
// for standalone "<" or ">" tokens. This is a quantifier decorator which
// indicates that the current token is to be repeated some number of times.
// It is used to quickly replicate data and allows the format to quickly
// generate large quantities of data.
//
// A sequence of tokens may be grouped together by enclosing them in a
// pair of "(" and ")" characters. Any standalone "<" or ">" tokens only take
// effect within the context of that group. A group with a "<" or ">" decorator
// is equivalent to having that as the first standalone token in the sequence.
// A repeator decorator on a group causes that group to be executed the
// specified number of times.
//
// If the total bit-stream does not end on a byte-aligned edge, then the stream
// will automatically be padded up to the nearest byte with 0 bits.
//
// Example BitGen file:
//	<<< # DEFLATE uses LE bit-packing order
//
//	( # Raw blocks
//		< 0 00 0*5                 # Non-last, raw block, padding
//		< H16:0004 H16:fffb        # RawSize: 4
//		X:deadcafe                 # Raw data
//	)*2
//
//	( # Dynamic block
//		< 1 10                     # Last, dynamic block
//		< D5:1 D5:0 D4:15          # HLit: 258, HDist: 1, HCLen: 19
//		< 000*3 001 000*13 001 000 # HCLens: {0:1, 1:1}
//		> 0*256 1*2                # HLits: {256:1, 257:1}
//		> 0                        # HDists: {}
//		> 1 0                      # Use invalid HDist code 0
//	)
//
// Generated output stream (in hexadecimal):
//	000400fbffdeadcafe000400fbffdeadcafe0de0010400000000100000000000
//	0000000000000000000000000000000000000000000000000000002c
func DecodeBitGen(s string) ([]byte, error) {
	packMode, toks, err := parse(s)
	if err != nil {
		return nil, err
	}

	bb := bitBuffer{packMode: packMode}
	if err := bb.Process(toks); err != nil {
		return nil, err
	}
	return bb.Bytes(), nil
}

type (
	token interface {
		Token()
	}

	basicToken struct {
		token
		order  byte
		repeat int
	}
	orderToken struct {
		*basicToken
	}
	bitsToken struct {
		*basicToken
		value  uint64
		length uint
	}
	bytesToken struct {
		*basicToken
		value []byte
	}
	groupToken struct {
		*basicToken
		tokens []token
	}
)

type parser struct {
	in  string
	rd  strings.Reader // For use with Fscanf only
	err error
}

func parse(in string) (order byte, toks []token, err error) {
	p := parser{in: in}
	defer func() {
		if ex := recover(); ex != nil {
			var ok bool
			if err, ok = ex.(error); ok {
				if err != io.ErrUnexpectedEOF {
					err = fmt.Errorf("parse error (offset:%d): %v", len(in)-len(p.in), err)
				}
				return
			}
			panic(ex)
		}
	}()

	// Parse for the bit-packing order.
	p.parseIgnored()
	if !strings.HasPrefix(p.in, "<<<") && !strings.HasPrefix(p.in, ">>>") {
		return 0, nil, errors.New("missing bit-packing order")
	}
	order, p.in = p.in[0], p.in[3:]

	// Parse for all other tokens.
	p.checkDelimiter()
	toks = p.parseTokens()
	if len(p.in) > 0 {
		return 0, nil, errors.New("unmatched group terminator") // Must be ')'
	}
	return order, toks, nil
}

// This returns only when reaching a group terminator or EOF of the input.
func (p *parser) parseTokens() (toks []token) {
	for {
		var btok basicToken
		var tok token
		p.parseIgnored()
		btok.order = p.parseOrder()

		var c byte
		if len(p.in) > 0 {
			c = p.in[0]
		}
		p.rd.Reset(p.in)
		switch c {
		case '0', '1':
			v := bitsToken{basicToken: &btok}
			_, p.err = fmt.Fscanf(&p.rd, "%b", &v.value)
			v.length = uint(len(p.in) - p.rd.Len())
			tok = v
		case 'D':
			v := bitsToken{basicToken: &btok}
			_, p.err = fmt.Fscanf(&p.rd, "D%d:%d", &v.length, &v.value)
			tok = v
		case 'H':
			v := bitsToken{basicToken: &btok}
			_, p.err = fmt.Fscanf(&p.rd, "H%d:%x", &v.length, &v.value)
			tok = v
		case 'X':
			v := bytesToken{basicToken: &btok}
			_, p.err = fmt.Fscanf(&p.rd, "X:%x", &v.value)
			tok = v
		case '"':
			v := bytesToken{basicToken: &btok}
			_, p.err = fmt.Fscanf(&p.rd, "%q", &v.value)
			tok = v
		case '(':
			v := groupToken{basicToken: &btok}
			p.in = p.in[1:]
			v.tokens = p.parseTokens() // Returns only when len(p.in) == 0 or p.in == ')'
			p.checkLength()
			p.in = p.in[1:]
			p.rd.Reset(p.in) // Avoid messing up the advancement logic below
			tok = v
		}
		p.checkError()
		p.in = p.in[len(p.in)-p.rd.Len():]

		if _, ok := tok.(bytesToken); ok && btok.order != 0 {
			p.err = errors.New("order cannot be set on bytes value")
			p.checkError()
		}
		if tok != nil {
			btok.repeat = p.parseRepeat()
		} else if btok.order != 0 {
			tok = orderToken{&btok}
		}
		p.checkDelimiter()
		toks = append(toks, tok)
		if c == ')' || len(p.in) == 0 {
			return toks
		}
	}
}

// parseIgnored skips past all preceding whitespace and comments.
func (p *parser) parseIgnored() {
	for len(p.in) > 0 && (unicode.IsSpace(rune(p.in[0])) || p.in[0] == '#') {
		if p.in[0] != '#' {
			p.in = p.in[1:]
		} else if i := strings.IndexByte(p.in, '\n'); i >= 0 {
			p.in = p.in[i:]
		} else {
			p.in = ""
		}
	}
}

// parseOrder parses out an optional order value.
func (p *parser) parseOrder() (order byte) {
	if len(p.in) > 0 && (p.in[0] == '<' || p.in[0] == '>') {
		order, p.in = p.in[0], p.in[1:]
	}
	return order
}

// parseOrder parses out an repeater value.
func (p *parser) parseRepeat() (n int) {
	if len(p.in) > 0 && p.in[0] == '*' {
		p.rd.Reset(p.in)
		_, p.err = fmt.Fscanf(&p.rd, "*%d", &n)
		p.checkError()
		p.in = p.in[len(p.in)-p.rd.Len():]
		return n
	}
	return 1
}

func (p *parser) checkLength() {
	if len(p.in) == 0 {
		panic(io.ErrUnexpectedEOF)
	}
}
func (p *parser) checkDelimiter() {
	if len(p.in) > 0 && !(unicode.IsSpace(rune(p.in[0])) || p.in[0] == '#' || p.in[0] == ')') {
		panic(fmt.Errorf("unexpected character: %c", p.in[0]))
	}
}
func (p *parser) checkError() {
	if p.err != nil {
		panic(p.err)
	}
}

type bitBuffer struct {
	packMode  byte
	parseMode byte

	b []byte
	m byte
}

func (b *bitBuffer) Process(toks []token) error {
	for _, t := range toks {
		switch t := t.(type) {
		case orderToken:
			b.parseMode = t.order
		case bitsToken:
			if t.length > 64 || t.value > uint64(1<<t.length)-1 {
				return fmt.Errorf("invalid bit value: D%d:%d", t.length, t.value)
			}
			if t.order == '>' || (t.order == 0 && b.parseMode == '>') {
				t.value = internal.ReverseUint64N(t.value, t.length)
			}
			for i := 0; i < t.repeat; i++ {
				b.WriteBits64(t.value, t.length)
			}
		case bytesToken:
			t.value = append([]byte(nil), t.value...) // Make copy
			if b.packMode == '>' {
				// Bytes tokens should not be affected by the bit-packing
				// order. Thus, if the order is reversed, we preemptively
				// reverse the bits knowing that it will reversed back to normal
				// in the final stage.
				for i, b := range t.value {
					t.value[i] = internal.ReverseLUT[b]
				}
			}
			if _, err := b.Write(bytes.Repeat(t.value, t.repeat)); err != nil {
				return err
			}
		case groupToken:
			saveMode := b.parseMode
			if t.order != 0 {
				b.parseMode = t.order
			}
			for i := 0; i < t.repeat; i++ {
				if err := b.Process(t.tokens); err != nil {
					return err
				}
			}
			b.parseMode = saveMode
		}
	}
	return nil
}

func (b *bitBuffer) Write(buf []byte) (int, error) {
	if b.m != 0x00 {
		return 0, errors.New("unaligned write")
	}
	b.b = append(b.b, buf...)
	return len(buf), nil
}

func (b *bitBuffer) WriteBits64(v uint64, n uint) {
	for i := uint(0); i < n; i++ {
		if b.m == 0x00 {
			b.m = 0x01
			b.b = append(b.b, 0x00)
		}
		if v&(1<<i) != 0 {
			b.b[len(b.b)-1] |= b.m
		}
		b.m <<= 1
	}
}

func (b *bitBuffer) Bytes() []byte {
	buf := append([]byte(nil), b.b...)
	if b.packMode == '>' {
		for i, b := range buf {
			buf[i] = internal.ReverseLUT[b]
		}
	}
	return buf
}
