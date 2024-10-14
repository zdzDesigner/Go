package main

import (
	"fmt"
	"regexp"
)

func main() {
	// 创建正则表达式
	re := regexp.MustCompile(`^@\/.*\.js$`)

	// 测试字符串
	tests := []string{
		"@/example.js",
		"@/path/to/file.js",
		"@/not/a/js/file.txt",
		"just/a/normal/file.js",
		"@/another/file.jsx",
	}

	// 检查每个测试字符串是否匹配
	for _, test := range tests {
		if re.MatchString(test) {
			fmt.Printf("%s matches\n", test)
		} else {
			fmt.Printf("%s does not match\n", test)
		}
	}
}
