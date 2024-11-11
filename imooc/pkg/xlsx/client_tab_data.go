package xlsx

import (
	"encoding/json"
	"fmt"

	"github.com/360EntSecGroup-Skylar/excelize"
)

type ClientData struct {
	ServiceName       string `json:"serviceName"`
	Office            string `json:"office"`
	Type              string `json:"type"`
	Routing           string `json:"routing"`
	Address           string `json:"address"`
	RoutingParameters string `json:"routingParameters"`
}

func ClientTabData() {
	f, err := excelize.OpenFile("pkg/xlsx/client_tab_data2.xlsx")
	if err != nil {
		println(err.Error())
		return
	}
	// 获取单个空格数据
	// cell := f.GetCellValue("Sheet1", "B1")
	// fmt.Println(cell)

	// 获取 Sheet 数据.
	rows := f.GetRows("Sheet1")

	fmt.Println(rows)
	fmt.Println(len(rows))
	data := make(map[string]ClientData)

	for _, row := range rows[1:] {
		var (
			tabName    = ""
			clientData = ClientData{}
		)
		for i, cell := range row {
			fmt.Println(i, cell)
			if i == 0 {
				tabName = cell
				clientData.ServiceName = cell
			}
			if i == 1 {
				clientData.Office = cell
			}
			if i == 2 {
				clientData.Type = cell
			}
			if i == 3 {
				clientData.Routing = cell
			}
			if i == 4 {
				clientData.Address = cell
			}
			if i == 5 {
				clientData.RoutingParameters = cell
			}

		}

		data[tabName] = clientData
	}
	// fmt.Println(data)
	if bt, err := json.Marshal(data); err == nil {
		fmt.Println(string(bt))
	}

}
