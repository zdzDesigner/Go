package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func main() {
	// base()
	// base1()
	// base2()
	// base3()
	// base4()
	// base5()
	// base6()
	// base7()
	// base8()
	// base9()
	// base10()
	// base11()
	// base12()
	// base13()
	// base14()
	// base14_1()
	// base14_2()
	// base15()
	// base16()
	// base17()
	base18()
}

func base17() {
	str := `goroutine 146 [running]:
	webhook/src/util/helper/errs.ErrStack(0xe552c0, 0x1853bb0, 0x1853bb0, 0x0)
		/home/zdz/Documents/Webhook/webhook/src/util/helper/errs/recover.go:25 +0x82
	webhook/src/middleware/cerror.Entry.func1(0xc023a31554e5dcd8, 0x52c1571b, 0x186cdc0, 0xc0003f2b00, 0xc000480070, 0xf, 0xc000390000, 0x4, 0xc000390005, 0x18, ...)
		/home/zdz/Documents/Webhook/webhook/src/middleware/cerror/index.go:87 +0x961
	panic(0xe552c0, 0x1853bb0)
		/home/zdz/Application/Go/go1.11.4/src/runtime/panic.go:513 +0x1b9
	webhook/src/routes/movienews.imgConv(0x10d6200, 0xc0001f0a00, 0xc00047d800, 0x5d, 0xc0003ce6c0, 0x0)
		/home/zdz/Documents/Webhook/webhook/src/routes/movienews/mixin.go:109 +0x2e2
	webhook/src/routes/movienews.Search(0x10e6300, 0xc0001f09c0)
		/home/zdz/Documents/Webhook/webhook/src/routes/movienews/search.go:35 +0x1ea
	webhook/src/util/wh.IHander.func1(0xc0003f2b00)
		/home/zdz/Documents/Webhook/webhook/src/util/wh/ctx.go:73 +0x67
	github.com/gin-gonic/gin.(*Context).Next(0xc0003f2b00)
		/home/zdz/go/pkg/mod/github.com/gin-gonic/gin@v1.3.0/context.go:108 +0x43
	webhook/src/routes/movienews.glob..func2.1(0xc0003f2b00)
		/home/zdz/Documents/Webhook/webhook/src/routes/movienews/search_test.go:26 +0x1c4
	github.com/gin-gonic/gin.(*Context).Next(0xc0003f2b00)
		/home/zdz/go/pkg/mod/github.com/gin-gonic/gin@v1.3.0/context.go:108 +0x43
	webhook/src/middleware/auth.WListRule(0xc0003f2b00)
		/home/zdz/Documents/Webhook/webhook/src/middleware/auth/wlist.go:33 +0x53
	github.com/gin-gonic/gin.(*Context).Next(0xc0003f2b00)
		/home/zdz/go/pkg/mod/github.com/gin-gonic/gin@v1.3.0/context.go:108 +0x43
	webhook/src/middleware/auth.WListContainer(0xc0003f2b00)
		/home/zdz/Documents/Webhook/webhook/src/middleware/auth/wlist.go:68 +0x2ef
	github.com/gin-gonic/gin.(*Context).Next(0xc0003f2b00)
		/home/zdz/go/pkg/mod/github.com/gin-gonic/gin@v1.3.0/context.go:108 +0x43
	webhook/src/middleware/auth.Dmprotl(0xc0003f2b00)
		/home/zdz/Documents/Webhook/webhook/src/middleware/auth/dmprotl.go:27 +0x39
	github.com/gin-gonic/gin.(*Context).Next(0xc0003f2b00)
		/h`

	val := regexp.MustCompile(`[\s\S]+runtime\/panic.go.+`).ReplaceAllString(str, "")
	val = regexp.MustCompile(`.+gin-gonic\/gin.+|.+\/src|\+(.+)`).ReplaceAllString(val, "")
	fmt.Println(val[0:300])
}

func regChain(reg string) {
	// a.replace(/[\s\S]+runtime\/panic.go.+/,'').replace(/.+gin-gonic\/gin.+|.+\/src|\+(.+)/g,'')
}

func base18() {
	val := regexp.MustCompile(`^.*/`).ReplaceAllString("./werwr/erwer.mp3", "")
	fmt.Println(val)
}

func base16() {
	val := regexp.MustCompile(`主播`).ReplaceAllString("主播杨卫&主播龙夜&主播任钦臣", "")
	fmt.Println(val)
}

func base15() {
	val := regexp.MustCompile(`.*-(\d{2})-(\d{2}).*`).ReplaceAllStringFunc("2021-05-06 sdd", func(str string) string {
		fmt.Println(str)
		return "$1"
	})
	fmt.Println(val)
	val2 := regexp.MustCompile(`.*-(\d{2})-(\d{2}).*`).FindAllStringSubmatch("2021-a-06 sdd", -1)
	fmt.Println(val2)
}

func base14_2() {
	val := regexp.MustCompile(`[（）\(\)]`).ReplaceAllString("今（手动阀s）天啊a(sdf啊，)sd(aw)cc", "")
	fmt.Println(val)
}

func base14_1() {
	val := regexp.MustCompile(`（.+?）|\(.+?\)`).ReplaceAllString("今（手动阀s）天啊a(sdf啊，)sd(aw)cc", "")
	fmt.Println(val)
}

func base14() {
	val := regexp.MustCompile(`\(.+?\)`).ReplaceAllString("今天啊a(sdf啊，)sd(aw)cc", "")
	fmt.Println(val)
}

func base13() {
	reg := regexp.MustCompile(`([a-z]+)$`)
	fmt.Println(reg.MatchString("舒缓惬意的asds"))
	fmt.Println(reg.MatchString("舒"))
	fmt.Println(reg.ReplaceAllString("舒缓惬意的自girl", "$0"))
	fmt.Println(reg.ReplaceAllString("舒缓惬意的自girl", "$1"))
	fmt.Println(reg.ReplaceAllString("舒缓惬意的自girl", "$2"))
	fmt.Println(reg.FindSubmatch([]byte("舒缓惬意的自girl")))
	fmt.Println(reg.FindStringSubmatch("舒缓惬意的自girl"))
	fmt.Println(reg.FindString("舒缓惬意的自girl"))
	fmt.Println(reg.FindString("TWEEgirl"))
}

func base12() {
	str := `<li id="audioItemBox_43002796" class="audio-item-box audio-item-box-search-mode-true" itemid="43002796" itembox="43002796" lastitem="" set-id="">
    <input type="hidden" id="itemInfoToken_audio_mp3_43002796" itemid="43002796" fuid="" ftype="audio_mp3" fmodel="play" rurl="43002796" extime="1604638800875" token="cd666ef4c9" cbk="callBackAudioFilePlay" rcbk="downloadAudioCallback">
<div class="audio-item-row audio-item-row1">

        <div class="title">
            <span class="title-name">
                    舒缓惬意的自然之音</span>
            </div>

        <div class="audio-file-info">
            <span class="quality" title="SQ无损品质 立体声  比特率1411k 采样率:44k 格式:wav 文件大小:56m">SQ无损品质</span><span class="split">|</span>
                <span class="down-num">
                    <i class="glyphicon glyphicon-download-alt download-icon"></i>
                        111</span>
            <span class="split">|</span>
                </div>

    </div>

    <div class="audio-item-row audio-item-row2">

        <div class="audio-wave-player">
        <span class="audio-player-btn"><i class="icon-play"></i></span></div>

        <div class="audio-right-box">
                <div class="audio-player-options">
                    <div class="dropdown speed" title="选择播放倍速">
                        <span id="dLabelSpeed_43002796" type="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
                            <span style="border: 1px solid #aaa;border-radius: 5px;padding: 1px 5px;">
                                <span class="speed-title">x1倍</span>
                                <span style="position: relative;top: 9px;font-size: 18px;line-height: 13px;">ˇ</span>
                            </span>

                        </span>
                        <ul class="dropdown-menu" aria-labelledby="dLabel">
                            <li speed="0.7">×0.7倍</li>
                            <li class="selected" speed="1">×1倍</li>
                            <li speed="1.5">×1.5倍</li>
                            <li speed="2">×2倍</li>
                            <li speed="2.5">×2.5倍</li>
                            <li speed="3">×3倍</li>
                        </ul>
                    </div>

                    <div class="setting" title="播放器设置">
                        <i class="icon-cog" style="color: #aaa"></i>
                        <div class="setting-body" style="display: none;">
                            <div style="width: 120px; height: 60px;">
                                <div id="change_engine_43002796" class="change-engine">
                                    <select class="play-engine">
                                        <option value="h5Play">h5播放器</option>
                                        <option value="flashPlay">flash播放器</option>
                                    </select>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="complain">
                        <span class="complain-btn btn btn-link" style="padding:0 6px;position: relative;top:-2px;" unit-type="item" unit-id="43002796" title="点击举报">
<i class="glyphicon glyphicon-alert"></i>
<span>举报</span>
</span></div>
                </div>

                <div class="audio-info-license">
                    <span title="Edie Elkinson" class="author">
                            作者：Edie Elkinson</span>
                        </div>
                </div>

        <span class="duration" title="音效长度" duration="330000">
            <i class="icon-time"></i>
            05:30</span>
        <div class="mark-bar mark-bar-row" id="mark-bar-item-43002796"><span style="padding-right: 10px;color:#006633;"><i class="icon-play"></i><span class="last-time">昨天</span></span></div>

    <div class="audio-download-box audio-download-box-copyright">
    <span class="btn btn-default audio-buy-btn" onclick="gei.buy.license.choose('item','43002796');">
    <i class="icon-shopping-cart"></i>
    购买授权
    </span>

    <input type="hidden" id="itemInfoToken_audio_mp3_down_43002796" itemid="43002796" fuid="" ftype="audio_mp3_down" fmodel="down" rurl="43002796" extime="1604638800888" token="4a7c056601" rcbk="downloadAudioCallback">
<span class="btn btn-default audio-prev-down-btn" onclick="itemFileDown(this)" itemid="43002796" item-down-type="audio_mp3_down">
    <i class="icon-download-alt"></i>
    下载试听
    </span>
    <span class="fav-btn btn btn-default" style="" fav="false" unit-type="item" unit-id="43002796" title="点击收藏">
<span id="fav-anim-bar-item-43002796" class="icon-star-empty mark-animation hide" style="position: absolute;"></span>
<i class="icon-star-empty"></i>
<span>收藏</span>
</span></div></div>

    <div class="audio-info-dir-group">
        &nbsp;</div>

    <div class="unittags-container">
        <span class="tags" title="舒缓">舒缓</span>
            <span class="tags" title="轻松">轻松</span>
            <span class="tags" title="宁静">宁静</span>
            <span class="tags" title="惬意">惬意</span>
            <span class="tags" title="纪录片">纪录片</span>
            <span class="tags" title="旅游">旅游</span>
            <span class="tags" title="新世纪">新世纪</span>
            <span class="tags" title="轻音乐">轻音乐</span>
            <span class="tags" title="钢琴">钢琴</span>
            <span class="tags" title="小提琴">小提琴</span>
            <span class="tags" title="轻快">轻快</span>
            <span class="tags" title="能量低">能量低</span>
            <span class="tags" title="平稳">平稳</span>
            </div>
</li>`
	reg := regexp.MustCompile(`[\s\S]*<span class="title-name">[\n\r]*(.+)</span>[\s\S]*`)
	fmt.Println(reg.ReplaceAllString(str, "$1"))

	reg2 := regexp.MustCompile(`[\s\S]*?(<span class="tags" title=".+">[\n\r]*(.+)</span>)?[\s\S]+?`)

	fmt.Println(reg2.ReplaceAllString(str, "$1"))
	fmt.Println(reg2.ReplaceAllString(str, "$2"))
	ret := reg2.ReplaceAllStringFunc(str, func(v string) string {
		if len(v) > 6 {
			reg3 := regexp.MustCompile(`^<span class="tags" title=".+">[\n\r]*(.+)</span>[\s\S]+?`)
			return reg3.ReplaceAllString(v, "$1") + "|"
		}
		return ""
	})
	fmt.Println(ret)

	// <span class="tags" title="舒缓">舒缓</span>
}

func base11() {
	reg := regexp.MustCompile(`[\s\S]*<span .*>[\n\r]*(.+)</span>[\s\S]*`)

	str := `
	<div class="title">
		<span class="title-name">
				  舒缓惬意的自然之音</span>
	</div>`
	fmt.Println(reg.ReplaceAllString(str, "$0"))
	fmt.Println(reg.ReplaceAllString(str, "$1"))
	fmt.Println(strings.Trim(reg.ReplaceAllString(str, "$1"), "\t"))
	fmt.Println(strings.Trim(strings.Trim(reg.ReplaceAllString(str, "$1"), "\t"), " "))

	str2 := `<span class="title-name">舒缓惬意的自然之音</span>`
	fmt.Println(reg.ReplaceAllString(str2, "$0"))
	fmt.Println(reg.ReplaceAllString(str2, "$1"))
}

func base10() {
	reg := regexp.MustCompile(`.*(\d{4}-\d{2}-\d{2})(.*)`)
	fmt.Println(reg.ReplaceAllString("2020-09-04中国大陆上映", "$0"))
	fmt.Println(reg.ReplaceAllString("2020-09-04中国大陆上映", "$1"))
	fmt.Println(reg.ReplaceAllString("2020-09-04中国大陆上映", "$2"))
}

func base9() {
	reg := regexp.MustCompile(`(,,,)|(,,)`)
	fmt.Println(reg.ReplaceAllString("今天,,,天气,今天,,天气,不错", ","))
}

func base8() {
	// reg := regexp.MustCompile(`\D+`)
	reg := regexp.MustCompile(`\d+$`)
	// reg := regexp.MustCompile(`^.+?(\d+)$`)
	// fmt.Println(reg.ReplaceAllString("张少佐-多情剑客无情剑1034", "$1"))
	// fmt.Println(reg.FindAllString("张少佐-多情剑客无情剑1034", -1))
	fmt.Println(reg.FindSubmatch([]byte("张少佐-多情剑客3无情剑44")))
	fmt.Println(reg.FindStringSubmatch("张少佐-多情剑客3无情剑44"))
}

func base7() {
	reg := regexp.MustCompile(`^.+?(\d+)$`)
	fmt.Println(reg.ReplaceAllString("张少佐-多情剑客无情剑", "$1"))
}

func base6() {
	reg := regexp.MustCompile(`[\pP]+`)
	fmt.Println(reg.ReplaceAllString("今天，天气,不错", "xx"))
}

func base5() {
	reg := regexp.MustCompile(`[,|，]`)
	fmt.Println(reg.ReplaceAllString("今天，天气,不错", "|"))
}

func base4() {
	str := `爆笑修仙：兔子必须死  第122章 防不胜防啊`
	reg := regexp.MustCompile(`^.*第\d+章(.*)$`)
	fmt.Println(strings.Trim(reg.ReplaceAllString(str, "$1"), " "))

	str2 := `爆笑修仙：兔子必须死  第122集 啦啦`
	reg2 := regexp.MustCompile(`^.*第\d+[章|集](.*)$`)
	fmt.Println(strings.Trim(reg2.ReplaceAllString(str2, "$1"), " "))

	str3 := `爆笑修仙：片花`
	reg3 := regexp.MustCompile(`^.*第\d+[章|集](.*)$`)
	fmt.Println(strings.Trim(reg3.ReplaceAllString(str3, "$1"), " "))

	str4 := `爆笑修仙：兔子必须死  第122集 啦啦`
	reg4 := regexp.MustCompile(`^.*第(\d+)[章|集]\s*(.*)$`)
	fmt.Println(strings.Trim(reg4.ReplaceAllString(str4, "$1-$2"), " "))

	str5 := `爆笑修仙：兔子必须死  第122集   啦啦`
	reg5 := regexp.MustCompile(`^.*第(\d+)[章|集]\s*(.*)$`)
	fmt.Println(reg5.FindAllStringSubmatch(str5, -1))
	ret := reg5.ReplaceAllStringFunc(str5, func(str string) string {
		fmt.Println(str)
		return ""
	})
	fmt.Println(ret)
}

func base3() {
	str := `Host: 127.0.0.1:8000
User-Agent: curl/7.58.0
Accept: */*
Proxy:183.232.231.172
	`
	reg := regexp.MustCompile(`.*Proxy:(.*)\n.*`)
	// reg2 := regexp.MustCompile(`(?msU)Proxy:(.*)\n.*`)

	// proxy := reg.ReplaceAllString(str, "$1")
	// fmt.Printf("--%s--", proxy)

	fmt.Println(reg.FindAllStringSubmatch(str, -1)[0][1])

	// fmt.Printf("--%s--", strings.Replace(str, proxy, "", -1))
}

func base() {
	str := errors.New("=wh=StatusCode:304=wh=;sdsdfsdfsd306")

	reg := regexp.MustCompile(`^=wh=StatusCode:([0-9]{3})=wh=;.*`)
	fmt.Println(reg.ReplaceAllString(str.Error(), "$1"))
}

func base2() {
	str := `<audio id="xs" preload="auto"><source id="mySource" src="http://mobile.ximalaya.com/mobile/redirect/free/play/193203442/0"></audio>`
	reg := regexp.MustCompile(`.*\ssrc="(.*)".*`)
	fmt.Println(reg.ReplaceAllString(str, "$1"))
}

func base1() {
	reg := regexp.MustCompile(fmt.Sprintf(`^[0-9]{%d}$`, 5))
	fmt.Println(reg.MatchString("23433"))
}
