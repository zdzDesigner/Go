package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

func main() {
	// base()
	// base2()
	// base3()
	base4()
}
func base4() {
	val := float64(3.22) * float64(4)
	fmt.Println(val)
	fmt.Println(int(val))

}

func base3() {
	fmt.Println(toFixed(1.2345678, 0)) // 1
	fmt.Println(toFixed(1.2345678, 1)) // 1.2
	fmt.Println(toFixed(1.2345678, 2)) // 1.23
	fmt.Println(toFixed(1.2345678, 3)) // 1.235 (rounded up)

	fmt.Println(convUnit(0.235))
	fmt.Println(convUnit(0.))
	fmt.Println(convUnit(0))
	fmt.Println(convUnit(20.0))
	fmt.Println(convUnit(2023))
	fmt.Println(convUnit(20239))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func base() {
	rand.Intn(1)

}

func base2() {
	// x := 12.3456
	x := 0.002
	fmt.Println(math.Floor(x*100) / 100) // 12.34 (round down)
	fmt.Println(math.Round(x*100) / 100) // 12.35 (round to nearest)
	fmt.Println(math.Ceil(x*100) / 100)  // 12.35 (round up)
}

func convUnit(val float64) string {
	val = math.Round(val*100) / 100
	if val > 100000000 {
		return fmt.Sprintf("%.2f亿", val/100000000)
	}
	if val > 10000 {
		return fmt.Sprintf("%.2f万", val/10000)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", val), "0"), ".")
}
