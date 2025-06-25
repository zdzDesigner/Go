package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

func main() {
	// atoi()
	atof()
	// toLocaleLowerCase()
	// join()
	// format()
	// insertPoint()
	// containesEmpty()
	// length()
	// ch()
	// makeASCIISet("。")
	// parse()
	// encode()
	// foreach()
	// toDouble()
	// trim()
	// toInt()
	// base()
	// _split()
	// fmtHander()
	// v := nil1()
	// if v == nil {
	// 	return
	// }
	// fmt.Println(*v)
}

type VV struct{}

func nil1() *VV {
	return nil
}

func fmtHander() {
	a := _fmtHander()
	fmt.Printf("%s", a)
}

func _fmtHander() interface{} {
	return "aaa"
}

func _split() {
	fmt.Println(strings.Split("", " "), strings.Split("", " ")[0])
	fmt.Println(strings.Split("aaa bb   ccc", " "))
	fmt.Println(strings.Split("aaa bb   ccc", " "))
}

func base() {
	str := "aaa"
	fmt.Println(strings.Split(str, ","))
	fmt.Println("join:", strings.Join([]string{}, "and") == "")
}

func toInt() {
	v, _ := strconv.Atoi("09")
	fmt.Println(v == 9)
}

func trim() {
	fmt.Println(strings.Trim("sdsd的   a", " "), len(strings.Trim("sdsd的   a", " ")))
	fmt.Println(strings.Trim("sdsd的   a", ""))
	fmt.Println(strings.Trim("  sdsd的   a   ", " "), len(strings.Trim("  sdsd的   a   ", " ")))
	fmt.Println(strings.Trim("s dsd 的   ", " "), strings.Trim("sdsd的   ", " "), len(strings.Trim("sdsd的   ", " ")), len("sdsd的"))
	fmt.Println("---", strings.Trim("  sdsd的   ", " ") == "sdsd的")
	// fmt.Println(strings.ReplaceAll("s dsd 的   ", " ", ""))
	fmt.Println(strings.TrimRight("sdsd的", "。"))
	fmt.Println(strings.TrimRight("sdsd。的", "。"))
	fmt.Println(strings.TrimRight("sdsd。的。", "。"))
	// fmt.Println(strings.TrimLeft("21sds4d5", ""))
	v := strings.TrimLeftFunc("2133 sd s4d5", func(str rune) bool {
		return regexp.MustCompile(`^(\d+)+`).MatchString(string(str))
	})
	fmt.Println("strings.TrimLeftFunc:", v, len(v))

	fmt.Println(strings.TrimPrefix("09", "0"))
	fmt.Println(strings.TrimPrefix("90", "0"))
	fmt.Println(strings.TrimLeft("09", "0"))
}

func toDouble() {
	i, err := strconv.ParseFloat("4.5", 64)
	fmt.Println(i, err)
}

func foreach() {
	str := "你好oo11,,xx"
	for _, s := range str {
		fmt.Println(string(s))
	}
}

func encode() {
	fmt.Println("•")
}

func length() {
	str := "12345678年后"
	fmt.Printf("%p::\n", &str)
	fmt.Println("len:", len(str))
	fmt.Println(str[0:1])
	fmt.Println(str[5:9])
	fmt.Println(str[2:])
	fmt.Println(str[8:])
	fmt.Println(str[9:])
	fmt.Println("str[:] :: ", str[:])
}

func makeASCIISet(chars string) (as [8]uint32, ok bool) {
	for i := 0; i < len(chars); i++ {
		c := chars[i]
		fmt.Println(c)
		if c >= utf8.RuneSelf {
			return as, false
		}
	}
	return as, true
}

func containesEmpty() {
	fmt.Println(strings.Contains("aa", ""))
	fmt.Println(strings.Contains("aa", "a"))
}

func insertPoint() {
	str := "91100"
	fmt.Println(len(str), str[0])
	for i, v := range str {
		fmt.Println(i, v)
	}
	bs := []byte(str)
	fmt.Println(bs[:len(bs)-1])
	fmt.Println(bs[len(bs)-1])
	fmt.Println(string(append(bs[:len(bs)-1], byte('.'), bs[len(bs)-1])))
	// bs[:len(bs)-1]...,
}

func atof() {
	fmt.Println(strconv.ParseFloat("23.3333", 64))
}

func atoi() {
	fmt.Println(strconv.Atoi("aa"))
	fmt.Println(strconv.Atoi("1"))
}

func toLocaleLowerCase() {
	fmt.Println(strings.ToLower("VV"))
}

func join() {
	fmt.Println(strings.Join([]string{}, ","))
	var (
		intentName = "intentName"
		keys       = []string{"value1", "value2"}
	)
	fmt.Println(strings.Join(append([]string{intentName}, keys...), "-"))
}

type model struct{}

func format() {
	a := "{\"dialogMode\":{\"ageRange\":\"adult\",\"display\":\"general\"},\"播放有声书-菊与刀\":{\"intentName\":\"播放有声书\",\"albumId\":\"34795\",\"source\":\"lrts\",\"sectionTop\":5,\"section\":\"5\",\"title\":\"菊与刀\"}}"
	a = "{\"dialogMode\":{\"ageRange\":\"general\",\"display\":\"general\"},\"播放有声书-天下\":{\"title\":\"天下 第3章\",\"intentName\":\"播放有声书\",\"albumId\":\"26939\",\"source\":\"lrts\",\"expireTime\":1583890830386,\"slot-title\":\"天下\",\"section\":\"3\",\"durationPlay\":\"10:52\"}}"
	fmt.Println(a)
	v := make(map[string]map[string]string)
	json.Unmarshal([]byte(a), &v)
	fmt.Println(v)
	fmt.Println(v["播放有声书-天下"])
	fmt.Println(v["播放有声书-天下"]["section"])
	fmt.Println(v["播放有声书-天下"]["expireTime"] == "")
}

type ReqBody struct {
	Name      string `json:"name"`
	Author    string `json:"author"`
	Announcer string `json:"announcer"`
	Section   string `json:"section"`
	Type      string `json:"type"`
	Recommend string `json:"recommend"`
	Continue  string `json:"continue"`
}

func parse() {
	str := `{
		"pinyin": "bo fang tian xia",
		"continue": "{\"dialogMode\":{\"ageRange\":\"general\",\"display\":\"general\"},\"播放有声书-天下\":{\"title\":\"天下 第3章\",\"intentName\":\"播放有声书\",\"albumId\":\"26939\",\"source\":\"lrts\",\"expireTime\":1583890830386,\"slot-title\":\"天下\",\"section\":\"3\",\"durationPlay\":\"10:52\"}}",
		"intent": "播放有声书",
		"duiWidget": "media",
		"name": "天下",
		"rec": "播放天下",
		"nlu": {
			"timestamp": 1583889631,
			"skillId": "2020031000000012",
			"input": "播放天下",
			"semantics": {
				"request": {
					"task": "有声书",
					"slotcount": 3,
					"slots": [{
						"value": "播放",
						"pos": [0, 1],
						"rawvalue": "播放",
						"name": "操作",
						"rawpinyin": "bo fang"
					}, {
						"value": "天下",
						"pos": [2, 3],
						"rawvalue": "天下",
						"name": "书名",
						"rawpinyin": "tian xia"
					}, {
						"value": "播放有声书",
						"name": "intent"
					}]
				}
			}
		}
	}`
	var reqbody ReqBody
	json.Unmarshal([]byte(str), &reqbody)

	fmt.Println(reqbody.Continue)

	v := make(map[string]map[string]interface{})
	if err := json.Unmarshal([]byte(reqbody.Continue), &v); err != nil {
		fmt.Println(err)
	}
	fmt.Println(v)
	fmt.Println(v["播放有声书-天下"])
	fmt.Println(v["播放有声书-天下"]["section"])
	fmt.Println(v["播放有声书-天下"]["expireTime"] == "")
	if reqbody.Name != "" {
		fmt.Println(v[fmt.Sprintf("播放有声书-%s", reqbody.Name)], v[fmt.Sprintf("播放有声书-%s", reqbody.Name)]["section"])
	}
}

func ch() {
	str := "st播放aa"
	// -05 11111111.mp3
	// -05 今今遥可.mp3
	str = "-05 今今遥可" // 12
	str = "-乱觉- 幸圈乌呆石剪布Mimo）mp3"
	str = "乱感觉 - 幸福圈（乌拉呆+石头剪子布+Mimo）.mp3"
	// newstr := regexp.MustCompile("[\u4e00-\u9fa5]").ReplaceAllString(str, "11")
	newstr := regexp.MustCompile("[\u4e00-\u9fa5]").ReplaceAllLiteralString(str, "11")

	fmt.Println(newstr, len(newstr), len(newstr))
}
