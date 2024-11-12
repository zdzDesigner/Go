package lib

import (
	"fmt"
	"testing"
)

func Test_Docx2md(t *testing.T) {
	// err := Docx2md("/home/zdz/temp/word/"+"jds.doc", false)
	// err := Docx2md("/home/zdz/temp/word/"+"tiao_wen.docx", false)
	err := Docx2md("/home/zdz/temp/word/"+"响应式框架下基于JSON配置的UI编辑器.docx", false)
	if err != nil {
		fmt.Println(err.Error())
	}
}
