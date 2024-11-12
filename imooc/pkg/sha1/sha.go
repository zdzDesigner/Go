package sha1

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	str = `{"topic":"nlu.input.text","refText":"how are you"}`
	bt  = []byte(str)
)

type Body struct {
}

func (b *Body) Read(p []byte) (n int, err error) {
	// return 0, errors.New("bbb")

	// make([]byte, len(bt))
	n = copy(p, bt)
	return n, nil
}

// func (b *Body) Close() error {
// 	return nil
// 	// return errors.New("aaa")
// }

func Entry() {
	Base()
	// testProductServer()
	// testTuyaServer()
	// testShai()
	// testShai2()
	// testQny()
	// testShai3()
}

func testProductServer() {
	deviceName := "0ddddeeeeeeeeeeee88888888260c8ab"
	nonce := "bf7c8674"
	productId := "278578090"
	timestamp := "1546059559999"
	v := deviceName + nonce + productId + timestamp
	fmt.Println(HMACSHA1([]byte(v), v))

	// url := "http://dds.dui.ai/dds/v2/test?productId=279597020&apikey=69380252cad8426c917efb7c0af3cc03"
	url := "http://127.0.0.1:9060/data?aa=bb"

	// req, err := http.NewRequest("POST", url, bytes.NewBufferString(`{"topic":"nlu.input.text","refText":"how are you"}`))
	req, err := http.NewRequest("POST", url, &Body{})
	fmt.Println(req, err)
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(bt))
	req.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(bt)
		return ioutil.NopCloser(r), nil
	}
	resp, err := http.DefaultClient.Do(req)

	fmt.Println(resp, err)
	fmt.Println(resp.ContentLength, err)
	fmt.Println(resp.Header, err)
	fmt.Println(resp.Request, err)
	fmt.Println(resp.Body, err)
	defer resp.Body.Close()
	for {
		time.Sleep(time.Second)

		b := make([]byte, 1000)
		i, err := resp.Body.Read(b)
		fmt.Println(string(b), i, err)

		// b, err := io.CopyN(&W{}, resp.Body, 20)
		// fmt.Println(b, err)

	}
	// req.Body.Read
	// req.GetBody()
	// rc, err := req.GetBody()
	// fmt.Println(rc, err)
}

// a1ggTv5day9.iot-as-mqtt.cn-shanghai.aliyuncs.com
func testShai() {
	deviceSecret := "6564a09e058359fcaea71f5970b1cc9f"
	clientId := "clientId" + "12345"
	deviceName := "deviceName" + "rzLeuuQE2IiKmOPKurKf"
	productKey := "productKey" + "a1ggTv5day9"
	val := HMACSHA1([]byte(deviceSecret), clientId+deviceName+productKey)
	fmt.Println(val)
	// rzLeuuQE2IiKmOPKurKf&a1ggTv5day9
	// c3636b36c454e01610b528e91b294bc4f9bdad76
}
func testShai2() {
	deviceSecret := "10681644e35fb1da1e955da786e0f6c2"
	clientId := "clientId" + "12345"
	deviceName := "deviceName" + "KenX6lS2KEJ0nW68C8I8"
	productKey := "productKey" + "a1ggTv5day9"
	val := HMACSHA1([]byte(deviceSecret), clientId+deviceName+productKey)
	fmt.Println(val)
	// KenX6lS2KEJ0nW68C8I8&a1ggTv5day9
	// 3abe4782d3fd8c49980b9564a251f1aa0bfe52b0
}

func testShai3() {
	// a1vdkFdMsf5.iot-as-mqtt.cn-shanghai.aliyuncs.com
	deviceSecret := "0f5e6d07d9a47cef4184e4911540810b"
	clientId := "clientId" + "12345"
	deviceName := "deviceName" + "gOk9xLr88FdRVFRF2FRK"
	productKey := "productKey" + "a1vdkFdMsf5"
	val := HMACSHA1([]byte(deviceSecret), clientId+deviceName+productKey)
	fmt.Println(val)
	// gOk9xLr88FdRVFRF2FRK&a1vdkFdMsf5
	// 8dff9bfc8527f7ebcb36a2ae0dd12a04f4283fce
}

func testQny() {

	val := fmt.Sprintf("http://qjbp8suzg.hd-bkt.clouddn.com/audio/mxqf.mp3?e=%d", (time.Now().UnixNano()/1000000000)+3600)
	secret := "Yd5Y4pyzPKBn9iskJY9-qo4xJHsHUuXWtZdiAQa1"
	sign := base64HMACSHA1(secret, val)
	access := "QDngCfFPXVt8wCdxwIGm3xCO7uWvZq0cz0lCnXIO"
	token := fmt.Sprintf("%s:%s", access, sign)
	fmt.Println("sign:", sign)
	fmt.Printf("%s&token=%s", val, token)
}

func HMACSHA1(keyStr []byte, value string) string {

	key := []byte(keyStr)
	mac := hmac.New(sha1.New, key)
	if _, err := mac.Write([]byte(value)); err != nil {
		return ""
	}
	return hex.EncodeToString(mac.Sum(nil))
	//进行base64编码
	// res := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	// return res
}

func base64HMACSHA1(keyStr, value string) string {

	key := []byte(keyStr)
	mac := hmac.New(sha1.New, key)
	if _, err := mac.Write([]byte(value)); err != nil {
		return ""
	}
	// valRaw := hex.EncodeToString(mac.Sum(nil))
	//进行base64编码
	res := Base64UrlSafeEncode(mac.Sum(nil))
	return res
	// base64 safe
	// res := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	// return Base64UrlSafeEncode([]byte(res))

}
func Base64UrlSafeEncode(source []byte) string {
	// Base64 Url Safe is the same as Base64 but does not contain '/' and '+' (replaced by '_' and '-') and trailing '=' are removed.
	bytearr := base64.StdEncoding.EncodeToString(source)
	safeurl := strings.Replace(string(bytearr), "/", "_", -1)
	safeurl = strings.Replace(safeurl, "+", "-", -1)
	safeurl = strings.Replace(safeurl, "=", "", -1)
	return safeurl
}

type W struct{}

func (w *W) Write(p []byte) (n int, err error) {
	fmt.Println(string(p))

	return len(p), nil
}

func testTuyaServer() {
	client_id := "1KAD46OrT9HafiKdsXeg"
	access_token := "3f4eda2bdec17232f67c0b188af3eec1"
	t := "1588925778000"
	secret := "4OHBOnWOqaEC1mWXOpVL3yV50s0qGSRC"
	// sign_token = HMAC-SHA256(client_id + t, secret).toUpperCase()
	sign := hmacSha256(client_id+t, secret)
	fmt.Println(sign)
	// sign_bns = HMAC-SHA256(client_id + access_token + t, secret).toUpperCase()
	sign2 := hmacSha256(client_id+access_token+t, secret)
	fmt.Println(sign2)
}

func hmacSha256(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
