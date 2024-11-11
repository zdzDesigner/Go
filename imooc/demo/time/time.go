package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	// base()
	// base1()
	base2()
	// base3()
	// timeHandler()
	// base4()
	// base5()
	// base6()
	// base7()
	// since()
	// isZero()
	// tim()
	// sub()
	// addDate()
}

func addDate() {
	fmt.Println(time.Now().AddDate(0, 0, -1).Format("20060102"))
	fmt.Println(time.Now().AddDate(0, 0, 1).Format("20060102"))
}

func sub() {
	var tstart, tend, tstart1, tend1 time.Time
	tend = time.Now()
	fmt.Println(tstart.Sub(tend))
	fmt.Println("++++", tend.Sub(tstart).Seconds())
	fmt.Println("++++", tend.UnixNano())
	fmt.Println(tstart.Sub(tend).Seconds())
	// fmt.Println(tstart.Sub(tend).Microseconds())
	fmt.Println(float64(tstart.Sub(tend)))
	fmt.Println(float64(tstart.Sub(tend).Nanoseconds()) / 1000000)
	fmt.Println(float64(-1601013355875.000300))

	fmt.Println(tstart1.Sub(tend1))
}

func tim() {
	y, m, d := time.Now().Date()
	fmt.Println(y, int(m), d)
}

func isZero() {
	tim := time.Now().Add(time.Second * 10)
	fmt.Println((time.Time{}).IsZero())
	for {
		time.Sleep(time.Second)
		fmt.Println("---", tim.IsZero())
		if tim.IsZero() {
			break
		}

	}
	fmt.Println("zero after")
}

func base() {
	fmt.Println(time.Now().Year())
	fmt.Println(time.Now().YearDay())
}

func base1() {
	fmt.Println(time.Now().Format("2006-01-02T15:04:05Z"))
	fmt.Println(time.Parse("20060102", "20191210"))
	fmt.Println("Zone:========")
	fmt.Println(time.Now().UTC().Zone())
	fmt.Println(time.Parse("2006-01-02", "1970-01-01"))
}

func base2() {
	fmt.Println(time.Now().Second())
	fmt.Println(time.Now().UnixMilli(),int(time.Now().UnixMilli()))
	fmt.Println(time.Now().Unix())
	fmt.Println(time.Now().UnixNano())

	fmt.Println(time.Now().Nanosecond())
}

func timeHandler() {
	timer := time.Tick(time.Second * 5)
	for range timer {
		fmt.Println("timer ...")
	}
}

func base3() {
	// rand.Seed(time.Now().UnixNano())
	rand.Seed(time.Now().Unix())
	for i := 0; i < 10; i++ {
		x := rand.Intn(1000)
		fmt.Println(x)
		time.Sleep(time.Duration(x) * time.Millisecond)
	}
}

func base4() {
	// time.Local = time.FixedZone("CST", 3600*8)
	time.Local = time.FixedZone("UTC", 3600*8)
	// timestr := "2020-03-03T11:00:00+0800"
	timestr := "2020-03-04T18:00:00+0800"

	newTime, err := time.Parse("2006-01-02T15:04:05+0800", timestr)
	// 1583290800 1583316000
	fmt.Println(newTime, err)
	fmt.Println(newTime.Unix())
	fmt.Println(newTime.Unix(), 1583316000+(8*3600), newTime.Unix() == 1583316000+(8*3600))
	fmt.Println(newTime.Hour())
	fmt.Println(time.Now().Local().Unix())
	fmt.Println(time.Now().Local().Hour())
	fmt.Println(time.Now().Local().Unix() < newTime.Unix())
}

func base5() {
	// timestr := "2020-03-03T11:00:00+0800"
	timestr := "2020-03-04T11:00:00"

	newTime, err := time.Parse("2006-01-02T15:04:05", timestr)
	// 1583290800000
	fmt.Println(newTime, err)
	fmt.Println(newTime.Unix())
	fmt.Println(newTime.Hour())
	fmt.Println(time.Now().Local().Unix())
	fmt.Println(time.Now().Local().Hour())
	fmt.Println(time.Now().Local().Unix() < newTime.Unix())
}

func base6() {
	time.Local = time.FixedZone("CST", 3600*8)
	// time.Local = time.FixedZone("UTC", 3600*8)
	timestr := "2020-03-04T18:00:00+0800"

	newTime, err := time.Parse("2006-01-02T15:04:05+0800", timestr)
	// 1583290800 1583316000
	fmt.Println(newTime, err)
	fmt.Println(newTime.Unix())
	fmt.Println(newTime.Hour())
	fmt.Println(time.Now().Unix())
	fmt.Println(time.Now().Hour())
	fmt.Println(time.Now().Unix() < newTime.Unix()-8*3600)
}

// 时区
func base7() {
	now := time.Now()
	local1, err1 := time.LoadLocation("") // 等同于"UTC"
	if err1 != nil {
		fmt.Println(err1)
	}
	local2, err2 := time.LoadLocation("Local") // 服务器上设置的时区
	if err2 != nil {
		fmt.Println(err2)
	}
	local3, err3 := time.LoadLocation("America/Los_Angeles")
	if err3 != nil {
		fmt.Println(err3)
	}

	fmt.Println(now)
	fmt.Println(now.In(local1))
	fmt.Println(now.In(local2))
	fmt.Println(now.In(local3))

	timestr := "2020-03-04T18:00:00+0800"

	newTime, _ := time.Parse("2006-01-02T15:04:05+0800", timestr)
	fmt.Println(newTime.Hour())
	fmt.Println(newTime.In(local2).Hour())
	fmt.Println(newTime.Local().Hour())
	fmt.Println(newTime.UTC().Hour())
}

func since() {
	now := time.Now()
	time.Sleep(time.Second * 2)
	fmt.Println(time.Since(now))
}
