package util

import (
	"crypto/md5"
	"fmt"
)

func Md5(v string) string { return fmt.Sprintf("%x", md5.Sum([]byte(v))) }
