package cobweb

import (
"fmt"
"github.com/jinzhu/gorm"
_ "github.com/jinzhu/gorm/dialects/mysql"
"os"
)

func initDB() {
	fmt.Println(MySQLConfig.Url)
	db, err := gorm.Open("mysql", MySQLConfig.Url)
	if err != nil {
		fmt.Println( MySQLConfig, err.Error())
		os.Exit(-1)
	}

	//db.LogMode(true)
	db.DB().SetMaxIdleConns(MySQLConfig.MaxIdleConnections)
	db.DB().SetMaxOpenConns(MySQLConfig.MaxOpenConnections)
	db.DB().SetConnMaxLifetime(-1) //不重新利用，可以执行得更快

	//禁用复数表名
	db.SingularTable(true)

	//统一加前缀
	if MySQLConfig.TablePrefix != "" {
		gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
			return MySQLConfig.TablePrefix + defaultTableName
		}
	}

	DB = db
}

var DB * gorm.DB

func init() {
	initDB()
}

