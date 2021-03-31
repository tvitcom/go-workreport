package models

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var (
	PRODMODE bool
)
// Common pool prepare db connection
func InitDB(driverName, dataSourceName string) (*sql.DB, error) {
	conn, err := sql.Open(driverName, dataSourceName)
	log.Println("open main db conn")
	if err != nil {
		return nil, err
	}
	if err = conn.Ping(); err != nil {
		return nil, err
	}
	return conn, err
}
