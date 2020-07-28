package bloom

import (
	"encoding/binary"
	"reflect"
	"unsafe"
)

func str2Bytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func uint16ToBytes(num uint16) []byte {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, num)
	return data
}

func uint32ToBytes(num uint32) []byte {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, num)
	return data
}

func uint64ToBytes(num uint64) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, num)
	return data
}
