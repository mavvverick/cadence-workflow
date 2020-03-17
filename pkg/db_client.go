package pkg

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
)

// ConnectSQL creates a sql connection
func ConnectSQL() (*gorm.DB, error) {
	var dbVersion sqlResult
	host := "34.93.3.207"
	database := "prod"
	// host := "10.65.96.3"
	// database := "prod"
	port := 3306
	username := "theseus_prod"
	password := "AK6xgnkDwqKh2HdH"
	dbSource := fmt.Sprintf(
		"%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True",
		username,
		password,
		host,
		port,
		database,
	)
	if flag.Lookup("test.v") != nil {
		dbSource = os.Getenv("dbSourceTest")
	}
	db, err := gorm.Open("mysql", dbSource)
	db.DB().SetConnMaxLifetime(time.Minute * 5)
	db.DB().SetMaxIdleConns(0)
	db.DB().SetMaxOpenConns(10)
	//  db.LogMode(true)
	if err := db.Raw("SELECT version() as version").Scan(&dbVersion).Error; err != nil {
		return db, err
	}
	fmt.Println("Sql version ", dbVersion)
	return db, err
}
