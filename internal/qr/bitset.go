// Derived from github.com/skip2/go-qrcode/bitset (MIT License).
// Copyright (c) 2014 Tom Harwood. See THIRD_PARTY_NOTICES in the repo root.
//
// Original: https://github.com/skip2/go-qrcode
// Modifications: extracted into internal package, removed unused methods.

package qr

import "log"

// bitset stores an append-only array of bits.
type bitset struct {
	numBits int
	bits    []byte
}

func newBitset(v ...bool) *bitset {
	b := &bitset{numBits: 0, bits: make([]byte, 0)}
	b.appendBools(v...)
	return b
}

func cloneBitset(from *bitset) *bitset {
	return &bitset{numBits: from.numBits, bits: from.bits[:]}
}

func (b *bitset) substr(start int, end int) *bitset {
	if start > end || end > b.numBits {
		log.Panicf("Out of range start=%d end=%d numBits=%d", start, end, b.numBits)
	}
	result := newBitset()
	result.ensureCapacity(end - start)
	for i := start; i < end; i++ {
		if b.at(i) {
			result.bits[result.numBits/8] |= 0x80 >> uint(result.numBits%8)
		}
		result.numBits++
	}
	return result
}

func (b *bitset) appendBytes(data []byte) {
	for _, d := range data {
		b.appendByte(d, 8)
	}
}

func (b *bitset) appendByte(value byte, numBits int) {
	b.ensureCapacity(numBits)
	if numBits > 8 {
		log.Panicf("numBits %d out of range 0-8", numBits)
	}
	for i := numBits - 1; i >= 0; i-- {
		if value&(1<<uint(i)) != 0 {
			b.bits[b.numBits/8] |= 0x80 >> uint(b.numBits%8)
		}
		b.numBits++
	}
}

func (b *bitset) appendUint32(value uint32, numBits int) {
	b.ensureCapacity(numBits)
	if numBits > 32 {
		log.Panicf("numBits %d out of range 0-32", numBits)
	}
	for i := numBits - 1; i >= 0; i-- {
		if value&(1<<uint(i)) != 0 {
			b.bits[b.numBits/8] |= 0x80 >> uint(b.numBits%8)
		}
		b.numBits++
	}
}

func (b *bitset) ensureCapacity(numBits int) {
	numBits += b.numBits
	newNumBytes := numBits / 8
	if numBits%8 != 0 {
		newNumBytes++
	}
	if len(b.bits) >= newNumBytes {
		return
	}
	b.bits = append(b.bits, make([]byte, newNumBytes+2*len(b.bits))...)
}

func (b *bitset) append(other *bitset) {
	b.ensureCapacity(other.numBits)
	for i := 0; i < other.numBits; i++ {
		if other.at(i) {
			b.bits[b.numBits/8] |= 0x80 >> uint(b.numBits%8)
		}
		b.numBits++
	}
}

func (b *bitset) appendBools(bits ...bool) {
	b.ensureCapacity(len(bits))
	for _, v := range bits {
		if v {
			b.bits[b.numBits/8] |= 0x80 >> uint(b.numBits%8)
		}
		b.numBits++
	}
}

func (b *bitset) appendNumBools(num int, value bool) {
	for i := 0; i < num; i++ {
		b.appendBools(value)
	}
}

func (b *bitset) len() int { return b.numBits }

func (b *bitset) at(index int) bool {
	if index >= b.numBits {
		log.Panicf("Index %d out of range", index)
	}
	return (b.bits[index/8] & (0x80 >> byte(index%8))) != 0
}

func (b *bitset) byteAt(index int) byte {
	if index < 0 || index >= b.numBits {
		log.Panicf("Index %d out of range", index)
	}
	var result byte
	for i := index; i < index+8 && i < b.numBits; i++ {
		result <<= 1
		if b.at(i) {
			result |= 1
		}
	}
	return result
}
