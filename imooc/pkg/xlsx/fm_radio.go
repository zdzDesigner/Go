package xlsx

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
)

type FMItem struct {
	Key string
	Val []string
}

// FMRadioData 数据
func FMRadioData() {

	// tempdata := make(map[string][]string)
	data := make([]FMItem, 0)

	f, err := excelize.OpenFile("pkg/xlsx/fm_radio.xlsx")
	if err != nil {
		println(err.Error())
		return
	}
	// 获取单个空格数据
	// cell := f.GetCellValue("Sheet1", "B1")
	// fmt.Println(cell)

	// 获取 Sheet 数据.
	rows := f.GetRows("Sheet1")

	fmt.Println(len(rows))

	// var temp string
	// for _, row := range rows[1:] {
	// 	key := row[1]
	// 	val := row[0]
	// 	if key != "" {
	// 		temp = strings.Trim(key, " ")
	// 		tempdata[temp] = []string{}
	// 	}
	// 	tempdata[temp] = append(tempdata[temp], val)
	// }

	// for key, val := range tempdata {
	// 	data = append(data, FMItem{Key: key, Val: val})
	// }
	var item FMItem
	for _, row := range rows[1:] {
		key := row[1]
		val := row[0]
		if key != "" {
			item.Key = strings.Trim(key, " ")
			item.Val = make([]string, 0)
			data = append(data, item)
		}
		item.Val = append(item.Val, val)

	}

	if res, err := json.Marshal(data); err == nil {
		fmt.Println(string(res))
	}

}
