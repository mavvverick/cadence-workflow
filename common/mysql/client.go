package mysql


import (
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type DB struct {
	Client	*gorm.DB
}
// ConnectSQL to connect to mysql db
func (db *DB) ConnectSQL() {
	host := os.Getenv("SQL_HOST")
	user := os.Getenv("SQL_USER")
	pass := os.Getenv("SQL_PASS")
	port := os.Getenv("SQL_PORT")
	dbname := os.Getenv("SQL_DB")

	args := fmt.Sprintf(
		"%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True",
		user,
		pass,
		host,
		port,
		dbname,
	)
	db.Client, _ = gorm.Open("mysql", args)

	db.Client.DB().SetConnMaxLifetime(time.Minute * 5)
	db.Client.DB().SetMaxIdleConns(0)
	db.Client.DB().SetMaxOpenConns(10)
}
