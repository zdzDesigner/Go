package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

const PUB_KEY = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC5DK6Gz/KhcfcHUND5rxBqtbhkPaDgOGuACeWfDVM7nscTElVcB2acuHX2rBOjuWsjn1/ytILhV7amQ0KmkeH73994q89M/uXn83P3LN7gnYUwEJgI3FcxhHUIdvFrJVoo4QkTJV9lGzgWZUMpp4/l1Ur0WAJ+9lJhTRigRpM0nQIDAQAB"
const DATA = "username###timestamp"

// 公钥加密
func RsaEncrypt(data, keyBytes []byte) []byte {
	//解密pem格式的公钥
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("public key error"))
	}
	// 解析公钥
	// pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	pubInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)

	if err != nil {
		panic(err)
	}
	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)
	//加密
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		panic(err)
	}
	return ciphertext
}

func Entry() {
	// dir, _ := os.Getwd()
	// fmt.Println(dir)
	file, err := os.Open("/home/zdz/Documents/Try/Go/imooc/pkg/rsa/pub")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	bs, err := ioutil.ReadAll(file)
	fmt.Println(string(bs), err)
	fmt.Println(RsaEncrypt([]byte(DATA), []byte(PUB_KEY)))
}
