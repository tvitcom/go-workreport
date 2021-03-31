package models

import (
	"log"
    // "errors"
	"database/sql"
)

type (
	UserTable struct {
		Id_user  int    //` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
		Fio      string //` VARCHAR(64) NOT NULL,
		Email    string //` VARCHAR(64) NOT NULL,
		Password string //` VARCHAR(64) NULL DEFAULT NULL,
	}
	// iWebuser interface {
	// 	Test(conn *sql.DB, name string) error
	// 	DropUserById(conn *sql.DB, id int) error
	// 	GetUserIdByName(conn *sql.DB, name string) (id int, err error)
	// 	UpdateUserByName(conn *sql.DB, name string) (err error) //Or insert if not exist
	// }
)

func DropUserById(conn *sql.DB, id int) error {
	return nil
}

// Try find user and if not found then save as new user by name and return id_user
func GetUserIdByName(conn *sql.DB, name string) (int64) {
	var returnId int64

	// Test name is not empty
	if name == "" {
		return 0
	}

	q :="SELECT u.id_user FROM user u WHERE fio LIKE ? limit 1"
	stmt, err := conn.Prepare(q)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var id_user int64
	err = stmt.QueryRow("%" + name + "%").Scan(&id_user)

	if err == sql.ErrNoRows {
		log.Printf("no user with name %s", name)
		returnId = 0
	} else if err != nil {
		log.Fatal(err)
	} else {
		returnId = id_user
	}

	// If previous sql not finded then save record 
	if returnId == 0 {
		// If not then save into db
		result, err := conn.Exec("INSERT `user` SET `fio` = ?", name)
		if err != nil {
		    log.Fatal(err)
		}
		id_user, err = result.LastInsertId()
		if err != nil {
		    log.Fatal(err)
		}
	}
	returnId = id_user
	return returnId
}
