package cobweb

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

func initDB() {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("%scobweb.db", ExecPath)), &gorm.Config{
		Logger: nil,
	})
	if err != nil {
		fmt.Println("failed to connect database")
		os.Exit(0)
	}

	db.AutoMigrate(&Website{})

	DB = db
}

var DB *gorm.DB
