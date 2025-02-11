package config

import (
	"fmt"
	"os"

	"cameras/src/util/lang"
)

var (
	SERVER_PORT = os.Getenv("SERVER_PORT")
	PORT        = lang.IF3(SERVER_PORT == "", ":8088", SERVER_PORT)
	// PORT   = ":8084"
	DIR, _ = os.Getwd()
)

var (
	STORE_PATH     = fmt.Sprintf("%s/assets/", DIR)
	STORE_PATH_DIR = fmt.Sprintf("%s/assets", DIR)
	STORE_PATH_BOM = fmt.Sprintf("%s/assets/bom", DIR)
)

func init() {
	if Mysql.Addr == "" {
		fmt.Println("DIR:", STORE_PATH)
		Mysql.Username = "root"
		Mysql.Addr = "127.0.0.1:3306"
		Mysql.DBName = "battery"
	}

	if os.Getenv("ENV_ZDZ") == "prod" {
		PORT = ":6002"
	}
}

var Mysql = &struct {
	Username string
	Password string
	Addr     string
	DBName   string
}{
	Username: os.Getenv("MYSQL_USER"),
	Password: os.Getenv("MYSQL_PASSWORD"),
	Addr:     os.Getenv("MYSQL_HOST"),
	DBName:   os.Getenv("DATABASE_NAME"),
}
