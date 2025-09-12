package types

import "unsafe"

// UnsafeString DO NOT USE THIS, this makes the byte slice immutable, avoid copies
func UnsafeString(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

// UnsafeBytes returns a byte pointer without allocation.
func UnsafeBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
