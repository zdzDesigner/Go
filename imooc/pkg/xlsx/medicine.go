package xlsx

import (
	"encoding/json"
	"fmt"

	"golang.org/x/text/width"

	"github.com/360EntSecGroup-Skylar/excelize"
)

// MedicineData 数据
func MedicineData() {

	data := [][]string{}

	f, err := excelize.OpenFile("pkg/xlsx/data2011.xlsx")
	if err != nil {
		println(err.Error())
		return
	}
	// 获取单个空格数据
	// cell := f.GetCellValue("Sheet1", "B1")
	// fmt.Println(cell)

	// 获取 Sheet 数据.
	rows := f.GetRows("血压相关药物 （去除静脉）")
	fmt.Println(len(rows))
	for _, row := range rows[1:] {
		col := make([]string, 0, len(row))
		for _, cell := range row {
			str := width.Narrow.String(cell)
			// str1 := strings.Replace(str, "【", "[", -1)
			// str = strings.Replace(str1, "】", "]", -1)
			col = append(col, str)
		}
		data = append(data, col)
	}
	// for _, colCell := range rows[1] {
	// 	fmt.Println(colCell)
	// }
	if res, err := json.Marshal(data); err == nil {
		fmt.Println(string(res))
	}

}
