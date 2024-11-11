package main

import (
	"fmt"

	"golang.org/x/text/width"
)

func main() {
	s := `～２ 。，（）【】-1！@234567890abc１２３４５６７８９ａｂｃ`
	// 全角转半角
	fmt.Println(width.Narrow.String(s))
	// 半角转全角
	fmt.Println(width.Widen.String(s))
}
