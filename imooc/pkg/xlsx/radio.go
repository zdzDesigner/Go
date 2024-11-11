package xlsx

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
)

// RadioData 数据
func RadioData() {

	data := make(map[string][]string)

	f, err := excelize.OpenFile("pkg/xlsx/radio2.xlsx")
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

	for _, row := range rows[1:] {
		key := row[1]
		val := row[0]
		fmt.Println(val)
		if key != val {
			data[key] = strings.Split(val, "|")
		}
	}

	if res, err := json.Marshal(data); err == nil {
		fmt.Println(string(res))
	}

}
