package models

import (
	"log"
	"errors"
	// "strconv"
	"database/sql"
)

type (
)

func GetWorkreportLastAttemptId(conn *sql.DB, user_id int64) int {
	row := conn.QueryRow("SELECT MAX(wr.id_attempt) AS last_attempt FROM workreport wr WHERE wr.id_user = ?", user_id)
	var result int
	var t interface{}
	err := row.Scan(&t)
	if err != nil {
		log.Fatal(err)
	}
    switch t := t.(type) {
        case nil:
            result = 0
        case int:
            result = t
        default:
        }
	return result
}

func UpdateByYMDNums(conn *sql.DB, id_attempt, id_project, id_user, year, month_num, date_ms int , duration_m, task_name string) (int64, error) { //Or insert if not exist
	sql:="REPLACE workreport SET id_project, id_user, year, month_num, date_ms int , duration_m, task_name"
	stmtIns, err := conn.Prepare(sql) 
	if err != nil {
		log.Fatal(err)
	}
	defer stmtIns.Close()
	result, err := stmtIns.Exec(id_project, id_user, year, month_num, date_ms, duration_m, task_name) 
	if err != nil {
		log.Fatal(err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
	    log.Fatal(err)
	}
	return rows, nil
}

func GetSummaryDurationByMonth(conn *sql.DB, name string) (sumDuration string, err error) {
	return "0d 0h 0m", errors.New("Method is not ready!")
}
