package lang

import "unsafe"

func OrEmpty[T any](old T, now T) T {
	if *(*int)(unsafe.Pointer(&old)) == 0 {
		return now
	}
	return old
}
