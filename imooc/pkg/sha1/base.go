package sha1

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
)

func Base() string {
	guid := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	if _, err := io.WriteString(h, guid); err != nil {
		return ""
	}
	fmt.Println(h.Sum(nil), len(h.Sum(nil))) // length 20(byte), 40(hex)
	accept := make([]byte, 28)
	base64.StdEncoding.Encode(accept, h.Sum(nil))
	fmt.Println("ret:", string(accept))
	return string(accept)
}
