package cobweb

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

func initSqite() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("%scobweb.db", ExecPath)), &gorm.Config{
		Logger: nil,
	})
	if err != nil {
		fmt.Println("failed to connect database")
		os.Exit(0)
	}

	return db, nil
}
