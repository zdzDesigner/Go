package util

import (
	"io/ioutil"
	"os"
)

// LocalFileByte ..
func LocalFileByte(filepath string) ([]byte, error) {

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	bs, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
