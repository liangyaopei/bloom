package bloom

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/spaolacci/murmur3"
)

const (
	mod7       = 1<<3 - 1
	bitPerByte = 8
)

// Filter is the sturct of BloomFilter
// false positive error rate p approximately (1 - e^(-kn/m))^k
// probability of false positives decreases as m increases, and increases as n increases.
// k is number of hash function,
// m is the size of filter, n is the number of elements inserted
type Filter struct {
	lock       *sync.RWMutex
	concurrent bool

	m     uint64 // bit array of m bits, m will be ceiling to power of 2
	n     uint64 // number of inserted elements
	log2m uint64 // log_2 of m
	k     uint64 // the number of hash function
	keys  []byte // byte array to store hash value
}

// New is function of creating a bloom filter
// k is number of hash function,
// m is the size of filter
// race is sync or not
func New(size uint64, k uint64, race bool) *Filter {
	log2 := uint64(math.Ceil(math.Log2(float64(size))))
	filter := &Filter{
		m:          1 << log2,
		log2m:      log2,
		k:          k,
		keys:       make([]byte, 1<<log2),
		concurrent: race,
	}
	if filter.concurrent {
		filter.lock = &sync.RWMutex{}
	}
	return filter
}

// Add adds byte array to bloom filter
func (f *Filter) Add(data []byte) *Filter {
	if f.concurrent {
		f.lock.Lock()
		defer f.lock.Unlock()
	}
	h := baseHash(data)
	for i := uint64(0); i < f.k; i++ {
		loc := location(h, i)
		slot, mod := f.location(loc)
		f.keys[slot] |= 1 << mod
	}
	f.n++
	return f
}

// Test check if byte array may exist in bloom filter
func (f *Filter) Test(data []byte) bool {
	if f.concurrent {
		f.lock.RLock()
		defer f.lock.RUnlock()
	}
	h := baseHash(data)
	for i := uint64(0); i < f.k; i++ {
		loc := location(h, i)
		slot, mod := f.location(loc)
		if f.keys[slot]&(1<<mod) == 0 {
			return false
		}
	}
	return true
}

// AddString adds string to filter
func (f *Filter) AddString(s string) *Filter {
	data := str2Bytes(s)
	return f.Add(data)
}

// TestString if string may exist in filter
func (f *Filter) TestString(s string) bool {
	data := str2Bytes(s)
	return f.Test(data)
}

// AddUInt16 adds uint16 to filter
func (f *Filter) AddUInt16(num uint16) *Filter {
	data := uint16ToBytes(num)
	return f.Add(data)
}

// TestUInt16 checks if uint16 is in filter
func (f *Filter) TestUInt16(num uint16) bool {
	data := uint16ToBytes(num)
	return f.Test(data)
}

// AddUInt32 adds uint32 to filter
func (f *Filter) AddUInt32(num uint32) *Filter {
	data := uint32ToBytes(num)
	return f.Add(data)
}

// TestUInt32 checks if uint32 is in filter
func (f *Filter) TestUInt32(num uint32) bool {
	data := uint32ToBytes(num)
	return f.Test(data)
}

// AddUInt64 adds uint64 to filter
func (f *Filter) AddUInt64(num uint64) *Filter {
	data := uint64ToBytes(num)
	return f.Add(data)
}

// TestUInt64 checks if uint64 is in filter
func (f *Filter) TestUInt64(num uint64) bool {
	data := uint64ToBytes(num)
	return f.Test(data)
}

// AddBatch add data array
func (f *Filter) AddBatch(dataArr [][]byte) *Filter {
	if f.concurrent {
		f.lock.Lock()
		defer f.lock.Unlock()
	}
	for i := 0; i < len(dataArr); i++ {
		data := dataArr[i]
		h := baseHash(data)
		for i := uint64(0); i < f.k; i++ {
			loc := location(h, i)
			slot, mod := f.location(loc)
			f.keys[slot] |= 1 << mod
		}
		f.n++
	}
	return f
}

// AddUint16Batch adds uint16 array
func (f *Filter) AddUint16Batch(numArr []uint16) *Filter {
	data := make([][]byte, 0, len(numArr))
	for i := 0; i < len(numArr); i++ {
		byteArr := uint16ToBytes(numArr[i])
		data = append(data, byteArr)
	}
	return f.AddBatch(data)
}

// AddUint32Batch adds uint32 array
func (f *Filter) AddUint32Batch(numArr []uint32) *Filter {
	data := make([][]byte, 0, len(numArr))
	for i := 0; i < len(numArr); i++ {
		byteArr := uint32ToBytes(numArr[i])
		data = append(data, byteArr)
	}
	return f.AddBatch(data)
}

// AddUint64Batch adds uint64 array
func (f *Filter) AddUin64Batch(numArr []uint64) *Filter {
	data := make([][]byte, 0, len(numArr))
	for i := 0; i < len(numArr); i++ {
		byteArr := uint64ToBytes(numArr[i])
		data = append(data, byteArr)
	}
	return f.AddBatch(data)
}

// location returns the bit position in byte array
// & (f.m - 1) is the quick way for mod operation
func (f *Filter) location(h uint64) (uint64, uint64) {
	slot := (h / bitPerByte) & (f.m - 1)
	mod := h & mod7
	return slot, mod
}

// location returns the ith hashed location using the four base hash values
func location(h []uint64, i uint64) uint64 {
	// return h[ii%2] + ii*h[2+(((ii+(ii%2))%4)/2)]
	return h[i&1] + i*h[2+(((i+(i&1))&3)/2)]
}

// baseHash returns the murmur3 128-bit hash
func baseHash(data []byte) []uint64 {
	a1 := []byte{1} // to grab another bit of data
	hasher := murmur3.New128()
	hasher.Write(data) // #nosec
	v1, v2 := hasher.Sum128()
	hasher.Write(a1) // #nosec
	v3, v4 := hasher.Sum128()
	return []uint64{
		v1, v2, v3, v4,
	}
}

// Reset reset the bits to zero used in filter
func (f *Filter) Reset() {
	if f.concurrent {
		f.lock.Lock()
		defer f.lock.Unlock()
	}
	for i := 0; i < len(f.keys); i++ {
		f.keys[i] &= 0
	}
	f.n = 0
}

// MergeInPlace merges another filter into current one
func (f *Filter) MergeInPlace(g *Filter) error {
	if f.m != g.m {
		return fmt.Errorf("m's don't match: %d != %d", f.m, g.m)
	}

	if f.k != g.k {
		return fmt.Errorf("k's don't match: %d != %d", f.m, g.m)
	}
	if g.concurrent {
		return errors.New("merging concurrent filter is not support")
	}

	if f.concurrent {
		f.lock.Lock()
		defer f.lock.Unlock()
	}
	for i := 0; i < len(f.keys); i++ {
		f.keys[i] |= g.keys[i]
	}
	return nil
}

// Cap return the size of bits
func (f *Filter) Cap() uint64 {
	if f.concurrent {
		f.lock.RLock()
		defer f.lock.RUnlock()
	}
	return f.m
}

// KeySize return  count of inserted element
func (f *Filter) KeySize() uint64 {
	if f.concurrent {
		f.lock.RLock()
		defer f.lock.RUnlock()
	}
	return f.n
}

// FalsePositiveRate returns (1 - e^(-kn/m))^k
func (f *Filter) FalsePositiveRate() float64 {
	if f.concurrent {
		f.lock.RLock()
		defer f.lock.RUnlock()
	}
	expoInner := -(float64)(f.k*f.n) / float64(f.m)
	rate := math.Pow(1-math.Pow(math.E, expoInner), float64(f.k))
	return rate
}
