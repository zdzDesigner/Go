package kpi

import (
	"fmt"
)

func Entry() {
	total := 2590041612
	timeout := consumeHandler()
	warn := warnHandler()
	fmt.Printf("%s 占比 %.4f%%\n", "总稳定性占比", float64((total-timeout-warn)*100)/float64(total))

}

func consumeHandler() int {
	fmt.Println("耗时占比:")
	var (
		consume = []map[string]int{
			{"0~500": 2184903789 + 43547590},
			{"500~1000": 2594182 + 358538300},
			{"1000~1500": 211495},
			{"1500~2000": 44169 + 135440},
			{"2000~2500": 11840},
			{"2500~3000": 5476 + 12212},
			{"3000~3500": 9801},
			{"3500~4000": 14138},
			{"4000~4500": 5569},
			{"4500~5000": 3649},
			{"5000+": 4755 + 5805 + 2560},
		}
		sum     = 0
		timeout = 0
	)

	for _, item := range consume {
		for k, val := range item {
			sum += val
			if k == "5000+" {
				timeout = val
			}
		}
	}

	for _, item := range consume {
		// fmt.Println((item["val"]))
		for key, val := range item {
			fmt.Printf("耗时 %s 区间占比 %.4f%%\n", key, float64(val*100)/float64(sum))
		}
	}
	return timeout

}

func warnHandler() int {
	fmt.Println("告警占比:")
	var (
		sum     = 2590041612
		warn    = 0
		consume = []map[string]int{
			{"内容网关响应超时": 11358 + 516},
			{"内容网关故障": 2184 + 41},
			{"外部接口响应超时": 2366},
			{"外部接口故障": 17},
			{"CINFO响应超时": 1977},
			{"内部依赖服务故障": 709},
			{"内部依赖服务响应超时": 100 + 14 + 54},
			{"代码出现错误，导致当前服务panic": 5},
		}
	)

	for _, item := range consume {
		for _, val := range item {
			warn += val
		}
	}
	for _, item := range consume {
		// fmt.Println((item["val"]))
		for key, val := range item {
			fmt.Printf("%.4f%% 占比 - %s  \n", float64(val*100)/float64(sum), key)
		}
	}
	fmt.Printf("%s 占比 %.4f%%\n", "告警总占比", float64(warn*100)/float64(sum))
	return warn
}
