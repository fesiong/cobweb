package cobweb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type configData struct {
	DBConfig dbConfig `json:"db_config"`
}

var JsonData configData

var ExecPath string

func initJSON() {
	sep := string(os.PathSeparator)
	//root := filepath.Dir(os.Args[0])
	//ExecPath, _ = filepath.Abs(root)
	ExecPath, _ = os.Executable()
	baseName := filepath.Base(ExecPath)
	ExecPath = filepath.Dir(ExecPath)
	if strings.Contains(baseName, "go_build") || strings.Contains(ExecPath, "go-build") || strings.Contains(ExecPath, "/l_/") || strings.Contains(baseName, "Test") || strings.Contains(ExecPath, "Temp") {
		ExecPath, _ = os.Getwd()
	}
	length := utf8.RuneCountInString(ExecPath)
	lastChar := ExecPath[length-1:]
	if lastChar != sep {
		ExecPath = ExecPath + sep
	}
	log.Println(ExecPath)
	rawConfig, err := ioutil.ReadFile(fmt.Sprintf("%sconfig.json", ExecPath))
	if err != nil {
		fmt.Println("Invalid Config: ", err.Error())
		os.Exit(-1)
	}
	if err := json.Unmarshal(rawConfig, &JsonData); err != nil {
		fmt.Println("Invalid Config: ", err.Error())
		os.Exit(-1)
	}
	// create data directory
	_, err = os.Stat(ExecPath + "data")
	if os.IsNotExist(err) {
		_ = os.Mkdir(ExecPath+"data", os.ModePerm)
	}
}

func WriteConfig() error {
	//将现有配置写回文件
	configFile, err := os.OpenFile(fmt.Sprintf("%sconfig.json", ExecPath), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer configFile.Close()

	buff := &bytes.Buffer{}

	buf, err := json.MarshalIndent(JsonData, "", "\t")
	if err != nil {
		return err
	}
	buff.Write(buf)

	_, err = io.Copy(configFile, buff)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	initJSON()
}
