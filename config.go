package cobweb

import (
"encoding/json"
"fmt"
"io/ioutil"
"os"
"regexp"
"unicode/utf8"
)

type configData struct {
	MySQL mySQLConfig
}

type mySQLConfig struct {
	Database           string
	User               string
	Password           string
	Host               string
	Port               int
	Charset            string
	MaxIdleConnections int
	MaxOpenConnections int
	TablePrefix        string
	Url                string
}

var ExecPath string

func initJSON() {
	sep := string(os.PathSeparator)
	ExecPath, _ = os.Getwd()
	length := utf8.RuneCountInString(ExecPath)
	lastChar := ExecPath[length-1:]
	if lastChar != sep {
		ExecPath = ExecPath + sep
	}

	bytes, err := ioutil.ReadFile(fmt.Sprintf("%sconfig.json", ExecPath))
	if err != nil {
		fmt.Println("ReadFile: ", err.Error())
		os.Exit(-1)
	}

	configStr := string(bytes[:])
	reg := regexp.MustCompile(`/\*.*\*/`)

	configStr = reg.ReplaceAllString(configStr, "")
	bytes = []byte(configStr)

	if err := json.Unmarshal(bytes, &jsonData); err != nil {
		fmt.Println("Invalid Config: ", err.Error())
		os.Exit(-1)
	}

	//load Mysql
	MySQLConfig = jsonData.MySQL
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		MySQLConfig.User, MySQLConfig.Password, MySQLConfig.Host, MySQLConfig.Port, MySQLConfig.Database, MySQLConfig.Charset)
	MySQLConfig.Url = url
}

var jsonData configData
var MySQLConfig mySQLConfig

func init() {
	initJSON()
}
