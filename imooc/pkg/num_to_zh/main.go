package main

import (
	"fmt"
	"strconv"
	"strings"
)

func main() {
	base()
	// base2()
}

func base2() {
	to := []string{"十", "百", "千", "万", "十", "百", "千", "亿", "十", "百", "千", "万", "十", "百", "千"}
	zh := []string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九"}
	ret := make([]string, 0, len(to))
	// src := "734221"
	// src := "700221"
	// src := "712001"
	// src := "32000700201"
	// src := "10032000700201"
	src := "10000000000000"
	for i, v := range src {
		i = len(src) - 2 - i
		u := ""
		if i >= 0 {
			u = to[i]
		}
		// 数值
		val, _ := strconv.Atoi(string(v))
		if val != 0 {
			ret = append(ret, fmt.Sprintf("%s%s", zh[val], u))
		} else {
			ret = append(ret, "零")
			if u == "亿" || u == "万" {
				ret = append(ret, u)
			}
		}
	}
	fmt.Println(ret)
	newret := make([]string, 0, len(ret))
	innter := []string{} // 暂存
	for _, item := range ret {
		if item == "零" {
			innter = append(innter, "零")
		} else {
			if len(innter) >= 1 && item != "亿" && item != "万" {
				newret = append(newret, "零")
			}
			newret = append(newret, item)
			innter = []string{}
		}
	}
	fmt.Println(newret)
	fmt.Println(strings.Join(newret, ""))
}

func base() {
	to := []string{"十", "百", "千", "万", "十", "百", "千", "亿", "十", "百", "千", "万", "十", "百", "千"}
	zh := []string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九"}
	ret := make([]string, 0, len(to))
	// src := "734221"
	// src := "700221"
	// src := "700201"
	// src := "32000700201"
	// src := "10032000700201"
	src := "10000000000000"
	for i, v := range src {
		i = len(src) - 2 - i
		u := ""
		if i >= 0 {
			u = to[i]
		}
		// 数值
		val, _ := strconv.Atoi(string(v))
		if val != 0 {
			if u != "亿" && u != "万" {
				ret = append(ret, fmt.Sprintf("%s%s", zh[val], u))
			} else {
				ret = append(ret, zh[val])
			}
		} else {
			ret = append(ret, "零")
		}
		if u == "亿" || u == "万" {
			ret = append(ret, u)
		}
	}
	fmt.Println(ret)
	newret := make([]string, 0, len(ret))
	innter := []string{} // 暂存
	for _, item := range ret {
		if item == "零" {
			innter = append(innter, "零")
		} else {
			if len(innter) >= 1 && item != "亿" && item != "万" {
				newret = append(newret, "零")
			}
			// fmt.Println(item)
			newret = append(newret, item)
			innter = []string{}
		}

	}
	res := strings.Join(newret, "")
	fmt.Println(newret)
	fmt.Println(res)
	fmt.Println(strings.Replace(res, "亿万", "亿", -1))
}
