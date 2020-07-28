package _bloom_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/liangyaopei/bloom"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func TestAddAndGet(t *testing.T) {
	filter := bloom.New(1024, 3, false)
	// chaining operation
	filter.Add([]byte("Hello")).
		AddString("World").
		AddUInt16(uint16(16)).
		AddUInt32(uint32(32)).
		AddUInt64(uint64(64)).
		AddUint16Batch([]uint16{1, 2, 3})

	t.Logf("Hello exist:%t", filter.Test([]byte("Hello")))
	t.Logf("World exist:%t", filter.TestString("World"))
	t.Logf("uint 16 exist:%t", filter.TestUInt16(uint16(16)))
	t.Logf("uint 32 exist:%t", filter.TestUInt32(uint32(32)))
	t.Logf("uint 64 exist:%t", filter.TestUInt64(uint64(64)))

	t.Logf("key exist:%t", filter.Test([]byte("key")))
	t.Logf("exist exist:%t", filter.TestString("exist"))
	t.Logf("uint 128 exist:%t", filter.TestUInt16(uint16(128)))
	t.Logf("uint 33 exist:%t", filter.TestUInt32(uint32(33)))
	t.Logf("uint 65 exist:%t", filter.TestUInt64(uint64(65)))

}

func TestAddAndGet2(t *testing.T) {
	dataSize := 1000000
	dataMap := make(map[string]struct{}, dataSize)
	stringLen := 30
	filter := bloom.New(uint64(dataSize), 3, false)
	for i := 0; i < dataSize; i++ {
		randStr := randStringRunes(stringLen)
		// add unique random string
		if _, ok := dataMap[randStr]; !ok {
			dataMap[randStr] = struct{}{}
			filter.Add([]byte(randStr))
		}
	}
	for k := range dataMap {
		exist := filter.Test([]byte(k))
		if !exist {
			t.Fatalf("key %s not exist", k)
		}
	}
}

// TestAddAndGetSync test concurrent write and read in filter
func TestAddAndGetSync(t *testing.T) {
	sizeData := 100000
	stringLen := 30
	parts := 10

	filter := bloom.New(uint64(sizeData), 3, true)
	// concurrent write and read
	fn := func(size int, wg *sync.WaitGroup) {
		defer wg.Done()
		m := make(map[string]struct{}, size)
		for i := 0; i < size; i++ {
			randStr := randStringRunes(stringLen)
			// add unique random string
			if _, ok := m[randStr]; !ok {
				m[randStr] = struct{}{}
				// write
				filter.AddString(randStr)
				// read
				exist := filter.TestString(randStr)
				if !exist {
					t.Errorf("key %s not exist", randStr)
				}
			}
		}
	}
	var waitGroup sync.WaitGroup
	for i := 0; i < parts; i++ {
		waitGroup.Add(1)
		go fn(sizeData/parts, &waitGroup)
	}
	waitGroup.Wait()
}

func TestFalsePositive(t *testing.T) {
	dataSize := 1000000
	dataNoSize := 100000
	dataMap := make(map[string]struct{}, dataSize)
	dataNoMap := make(map[string]struct{}, dataNoSize)
	stringLen := 30
	filter := bloom.New(uint64(dataSize), 3, false)

	for i := 0; i < dataSize; i++ {
		randStr := randStringRunes(stringLen)
		if _, ok := dataMap[randStr]; !ok {
			dataMap[randStr] = struct{}{}
			filter.AddString(randStr)
		}
	}
	for i := 0; i < dataNoSize; i++ {
		randStr := randStringRunes(stringLen)
		// add unique random string
		_, ok := dataMap[randStr]
		if !ok {
			dataNoMap[randStr] = struct{}{}
		}
	}
	falsePositiveCount := 0
	for k := range dataNoMap {
		exist := filter.TestString(k)
		if exist {
			falsePositiveCount++
		}
	}
	falsePositiveRatio := float64(falsePositiveCount) / float64(dataNoSize)
	t.Logf("false positive count:%d,false positive ratio:%f", falsePositiveCount, falsePositiveRatio)
}

//  go test ./... -bench=BenchmarkFilter_Add  -benchmem -run=^$ -count=5
func BenchmarkFilter_Add(b *testing.B) {
	b.StopTimer()
	dataTestSize := 100000
	dataTestMap := make(map[string]struct{}, dataTestSize)
	dataTestArr := make([]string, dataTestSize)
	stringLen := 100
	for i := 0; i < dataTestSize; i++ {
		randStr := randStringRunes(stringLen)
		dataTestMap[randStr] = struct{}{}
		dataTestArr = append(dataTestArr, randStr)
	}
	filter := bloom.New(uint64(dataTestSize), 3, false)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < len(dataTestArr); i++ {
		filter.Add([]byte(dataTestArr[i]))
	}
	//for i := 0; i < b.N; i++ {
	//	for i := 0; i < len(dataTestArr); i++ {
	//		filter.Add([]byte(dataTestArr[i]))
	//	}
	//	//for i := 0; i < len(dataTestArr); i++ {
	//	//	//res := filter.Test([]byte(dataTestArr[i]))
	//	//	_, _ = dataTestMap[dataTestArr[i]]
	//	//	//if res != ok {
	//	//	//	b.Errorf("fatal.res:%t,ok:%t", res, ok)
	//	//	//}
	//	//}
	//}
}
