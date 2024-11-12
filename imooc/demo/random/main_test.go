package main

import (
	"testing"
)

func BenchmarkRand(b *testing.B) {
	// num := 10
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Rand()
		// fmt.Sprintf("%d", num)
	}
}

func BenchmarkRand2(b *testing.B) {
	// num := 10
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Rand2()
		// fmt.Sprintf("%d", num)
	}
}
