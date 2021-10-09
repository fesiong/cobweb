package cobweb

import (
	"gorm.io/gorm"
	"log"
	"os"
)

var DB *gorm.DB

func initDB() {
	if JsonData.DBConfig.Adapter == "mysql" {
		db, err := InitMySQL()
		if err != nil {
			log.Println("init mysql error: ", err.Error())
			os.Exit(1)
		}
		DB = db
	} else {
		db, err := initSqite()
		if err != nil {
			log.Println("init sqlite error: ", err.Error())
			os.Exit(1)
		}
		DB = db
	}

	DB.AutoMigrate(&Website{})
}
