package models

import (
	"log"
	"database/sql"
)
type (
	SettingTable struct {
		Key string
		Value string
	}
)

func SaveSettingItem(conn *sql.DB, key, val string) (int64, error) {
	//if key not set - insert new record:
	// Or update:
	result, err := conn.Exec("REPLACE `setting` SET `value` = ? WHERE `key` = ?", val, key )
	if err != nil {
	    log.Fatal(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
	    log.Fatal(err)
	}
	return rows, nil
}