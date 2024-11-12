package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
)

// 定义结构体与 XML 标签相匹配
type Bookstore struct {
	XMLName xml.Name `xml:"bookstore"`
	Books   []Book   `xml:"book"`
}

type Book struct {
	Title  Title  `xml:"title"`
	Author string `xml:"author"`
	Price  string `xml:"price"`
}

type Title struct {
	Lang string `xml:"lang,attr"` // `,attr` 必须有(属性), 不指定就是<lang>标签了
	Text string `xml:",chardata"` //  XML 元素的文本内容, 没有`,`就是<chardata>标签了
}

func main() {
	// 读取 XML 文件内容
	xmlFile, err := os.ReadFile("./tpl.xml")
	if err != nil {
		log.Fatalf("Error reading XML file: %v", err)
	}

	// 解析 XML
	var bookstore Bookstore
	err = xml.Unmarshal(xmlFile, &bookstore)
	if err != nil {
		log.Fatalf("Error unmarshalling XML: %v", err)
	}

	// 输出解析后的数据
	for _, book := range bookstore.Books {
		fmt.Printf("Title: %s (Language: %s)\n", book.Title.Text, book.Title.Lang)
		fmt.Printf("Author: %s\n", book.Author)
		fmt.Printf("Price: $%s\n\n", book.Price)
	}
}
