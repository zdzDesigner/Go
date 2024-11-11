package xlsx

import (
	"encoding/json"
	"fmt"
	"imooc/util"
	"strings"

	"golang.org/x/text/width"

	"github.com/360EntSecGroup-Skylar/excelize"
)

// Item ..
type Item struct {
	Phones []string `json:"phones"`
	Times  []string `json:"times"`
	City   string   `json:"city"`
	Level  int      `json:"level"`
	Type   string   `json:"type"`
}

// MindData 数据
func MindData() {

	data := make(map[string]*Item)

	f, err := excelize.OpenFile("pkg/xlsx/xljk２.xlsx")
	if err != nil {
		println(err.Error())
		return
	}
	// 获取单个空格数据
	// cell := f.GetCellValue("Sheet1", "B1")
	// fmt.Println(cell)

	// 获取 Sheet 数据.
	rows := f.GetRows("数据")

	fmt.Println(len(rows))
	// phones := []string{}
	// times := []string{}
	// locations := []string{}
	// citys := []string{}
	areas := make([]string, 0)
	areaRepeat := make([]string, 0)
	for _, row := range rows[1:] {
		area := row[2]
		if util.Contains(areas, area) && !util.Contains(areaRepeat, area) {
			areaRepeat = append(areaRepeat, area)
		}
		if area != "" {
			areas = append(areas, area)
			// fmt.Println(area)
		}
	}
	fmt.Println("areaRepeat:", areaRepeat)

	pool := [][]string{}
	for j, row := range rows[1:] {
		// if j > 27 {
		// 	return
		// }
		col := make([]string, 0, len(row))
		province, city, area, phone, time, typer := "", "", "", "", "", ""
		for i, cell := range row {

			str := width.Narrow.String(cell)
			if i == 0 {
				province = str
			}
			if i == 1 {
				city = str
			}
			if i == 2 {
				area = str
			}

			if i == 3 {
				phone = str
				// str = "电话:" + str
			}

			if i == 4 {
				time = strings.Replace(str, "\n", "", -1)
				// phone := str[2]
				// str = "时间:" + strings.Replace(str, "\n", "", -1)
			}
			if i == 5 {
				typer = str
			}
			col = append(col, str)

		}

		if province+city+area != "" && j != 0 {
			if province != "" {
				fmt.Println("pool:", pool)
				convGroupData(data, pool, areaRepeat)
				pool = [][]string{}
				fmt.Println("------------------- ======================")
			}
		}

		pool = append(pool, []string{province, city, area, phone, time, typer})
	}
	// for _, colCell := range rows[1] {
	// 	fmt.Println(colCell)
	// }
	// if res, err := json.Marshal(data); err == nil {
	// 	fmt.Println(string(res))
	// }

}

func createData() {
	// data := `{
	// 	"北京":{"phones":[],"times":[],"type":2,"level":1}
	// 	"吴中区":{"phones":[],"times":[],"type":2,"level":3}
	// }`
}

func convGroupData(res map[string]*Item, data [][]string, areaRepeat []string) {
	var (
		province string
		city     string
		area     string
	)
	for _, v := range data {
		province1, city1, area1, phone, time, typer := v[0], v[1], v[2], v[3], v[4], v[5]

		if province1 != "" {
			province = province1
			city = ""
			area = ""
		}
		if city1 != "" {
			city = city1
			area = ""
		}
		if area1 != "" {
			area = area1
		}
		fmt.Println(province, city, area, phone, time, typer)
		// for _, phone := range strings.Split(phones, "\n") {}

		if area != "" {
			// fmt.Println("+++++++", areaRepeat, area)
			if util.Contains(areaRepeat, area) {
				validItem(res, city+"-"+area, phone, time, city, typer, 3)
			} else {
				validItem(res, area, phone, time, city, typer, 3)
				// validItem(res, area, phone, time, city, typer, 3)
			}
			continue
		}
		if city != "" && city != area {
			validItem(res, city, phone, time, city, typer, 2)
			continue
		}
		if province != "" && province != city && phone != "" && time != "" {
			validItem(res, province, phone, time, city, typer, 1)
			continue
		}

		// fmt.Println(i, province, city, area, phone, time, typer)

	}
	// fmt.Printf("%+v\n", res)
	if res, err := json.Marshal(res); err == nil {
		fmt.Println(string(res))
	}
	// return res
}

func validItem(pool map[string]*Item, key string, phone, time, city, typer string, level int) {
	if pool[key] == nil {
		pool[key] = &Item{Phones: []string{phone}, Times: []string{time}, City: city, Level: level, Type: typer}
	} else {
		pool[key].Times = append(pool[key].Times, time)
		pool[key].Phones = append(pool[key].Phones, phone)
	}
}
