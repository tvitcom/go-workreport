package models

import (
	"errors"
	"database/sql"
)


type (
	ProjectTable struct { //` (
		Id_project  int    //` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
		Id_manager  int    //` INT(11) NOT NULL,
		Name        string //` VARCHAR(45) NOT NULL,
		Description string //` VARCHAR(45) NULL DEFAULT NULL,
	}

	iProject interface {
		UpdateByName(conn *sql.DB, name string) (err error) // Or insert if new
		GetProjectIdByString(conn *sql.DB, projString string) (pId int, err error)
	}
)

func (ProjectTable) UpdateProjectByName(conn *sql.DB, name string) (err error) {
	return errors.New("Method is not ready!")
}

func (ProjectTable) GetProjectIdByName(conn *sql.DB, projString string) (pId int, err error) {
	return 0, errors.New("Method is not ready!")
}
