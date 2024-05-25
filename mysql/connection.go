package mysql

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func Connection() {

	db, err := sql.Open("mysql", "tushar:password@tcp(localhost:3306)/ambox")

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	logError(err)

	err = db.Ping()

	logError(err)

	log.Println("Db connection successfull")

}

func logError(err error) {

	if err != nil {
		log.Fatal("Error occurred")
		panic(err.Error())
	}
}
