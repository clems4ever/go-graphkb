package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

var (
	username     = "svc-criteokb"
	password     = "@1wQ.WBE93~r1D~b5ia~"
	host         = "security-exp01-pa4.central.criteo.preprod"
	databaseName = "criteokb"
)

func main() {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s)/%s", username, password,
		host, databaseName))
	if err != nil {
		logrus.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		logrus.Fatal(err)
	}

	_, err = tx.Exec("INSERT INTO test (value) VALUES (?)", uint64(12435887700123278845))
	if err != nil {
		tx.Rollback()
		logrus.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		logrus.Fatal(err)
	}
}
