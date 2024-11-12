package opencc

import (
	"fmt"
	"imooc/lib/opencc"
)

// Entry ..
func Entry() {
	cc, err := opencc.NewOpenCC("s2t")
	if err != nil {
		fmt.Println(err)
		return
	}
	nText, err := cc.ConvertText(`迪拜（阿拉伯语：دبي，英语：Dubai），是阿拉伯联合酋长国人口最多的城市，位于波斯湾东南海岸，迪拜也是组成阿联酋七个酋长国之一——迪拜酋长国的首都。`)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(nText)
}
