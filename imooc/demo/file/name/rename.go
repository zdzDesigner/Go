package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
)

func main() {
	// rename()
	guiview()
}

func rename() {
	dir := "/home/zdz/temp/like-resource/乱感觉-幸福圈（乌拉呆+石头剪子布+Mimo）.mp3"

	newdir := regexp.MustCompile(`\s`).ReplaceAllString(dir, "")
	newdir = regexp.MustCompile(`（`).ReplaceAllString(newdir, "")
	newdir = regexp.MustCompile(`）`).ReplaceAllString(newdir, "")
	err := os.Rename(dir, newdir)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func guiview() {
	str := "乱感觉-幸福圈（乌拉呆+石头剪子布+Mimo）.mp3"

	for _, ch := range bytes.Runes([]byte(str)) {
		fmt.Println(string(ch))
	}
}
