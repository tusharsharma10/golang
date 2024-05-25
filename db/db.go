package db

import (
	"database/sql"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"restapi/util"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	// this is necessary for specifying the driver when connecting to DB
	_ "github.com/go-sql-driver/mysql"
)

// DB instance
type DB struct {
	Username string
	Password string
	Host     string
	Port     string
	Dbase    string
	Err      error
	db       *sql.DB
	Dbx      *sqlx.DB
}

type MultiInsertHolder struct {
	Query string
	Data  []interface{}
}

func getDBPassword(prefix string) string {
	// check existence of hashed password
	hashedPassword := os.Getenv(prefix + "DBENCRYPTEDPASSWORD")

	if len(hashedPassword) == 0 {
		return os.Getenv(prefix + "DBPASSWORD")
	}

	response, err := util.DecryptWithRandomIV([]byte(os.Getenv("DB_ENCRYPTION_SECRET_KEY")), hashedPassword)
	if err != nil {
		log.Fatal("error decrypting password %s", err)
	}

	return string(response)
}

// Conn : Initiation function
// Use instance := db.Conn()
// prefixes is used here so that you don't always have to specify an empty string
// it is just assumed
func Conn(env string, utf8 bool, maxOpenConn int, maxIdleConn int, prefixes ...string) *DB {
	instance := &DB{}

	prefix := ""

	if len(prefixes) > 0 {
		prefix = prefixes[0] + "_"
	}

	// load env if not already loaded

	if len(os.Getenv("DBUSER")) == 0 {
		_, b, _, _ := runtime.Caller(0)
		basepath := filepath.Dir(b)
		ap := path.Join(basepath, "../../config", env)

		if err := godotenv.Load(ap); err != nil {
			log.Fatalf("%s", err)
		}
	}

	instance.Username = os.Getenv(prefix + "DBUSER")
	instance.Dbase = os.Getenv(prefix + "DBNAME")
	instance.Host = os.Getenv(prefix + "DBHOST")
	instance.Port = os.Getenv(prefix + "DBPORT")
	instance.Password = getDBPassword(prefix)

	mysqlConnString := instance.Username + ":" + instance.Password + "@tcp(" + instance.Host + ":" + instance.Port + ")/" + instance.Dbase

	mysqlConnString += "?charset=utf8mb4&collation=utf8mb4_unicode_ci"

	instance.db, instance.Err = sql.Open("mysql", mysqlConnString)
	if instance.Err == nil {

		maxOpenConnections, errMaxOpen := strconv.Atoi(os.Getenv("MAXCONNECT"))
		maxIdleConnections, errMaxIdle := strconv.Atoi(os.Getenv("MAXIDLECONNECT"))

		if errMaxOpen != nil {
			maxOpenConnections = 10
		}

		if errMaxIdle != nil {
			maxIdleConnections = 3
		}

		if maxOpenConn >= 0 {
			maxOpenConnections = maxOpenConn
		}

		if maxIdleConn >= 0 {
			maxIdleConnections = maxIdleConn
		}

		instance.db.SetMaxOpenConns(maxOpenConnections)
		instance.db.SetMaxIdleConns(maxIdleConnections)
		instance.db.SetConnMaxLifetime(time.Hour)
	} else {
		log.Fatalf("\nError connecting to DB with connection string = %s", mysqlConnString)
		log.Fatalf(instance.Err.Error())
		panic("\nCouldn't connect to DB\n")
	}

	instance.Dbx = sqlx.NewDb(instance.db, "mysql")

	return instance
}

// Fetch : Function that fetches through and returns first result
//
//	row := instance.Fetch("SELECT * FROM user WHERE Id = :Id", map[string]string{
//																		'Id' : 2
//																	})

// func (db *DB) Fetch(query string, args ...interface{}) map[string]string {
// 	dbResponse, err := gosqlcrud.QueryToMaps(db.db, query, args...)
// 	if err != nil {
// 		db.Err = err
// 		return nil
// 	}

// 	if dbResponse == nil || len(dbResponse) == 0 {
// 		return nil
// 	}

// 	return dbResponse[0]
// }

// FetchOrError : Function that fetches through and returns first result and error if any

// func (db *DB) FetchOrError(query string, args ...interface{}) (map[string]string, error) {
// 	dbResponse, err := gosqlcrud.QueryToMaps(db.db, query, args...)

// 	if dbResponse == nil || len(dbResponse) == 0 {
// 		return nil, err
// 	}

// 	return dbResponse[0], err
// }

// FetchAll : Function that fetches through and returns all the results
//
//	rows := instance.FetchAll("SELECT * FROM user WHERE Id = :Id", map[string]string{
//																		'Id' : 2
//																	})

// func (db *DB) FetchAll(query string, args ...interface{}) []map[string]string {
// 	dbResponse, err := gosqlcrud.QueryToMaps(db.db, query, args...)
// 	if err != nil {
// 		db.Err = err
// 		return nil
// 	}

// 	return dbResponse
// }

// Execute : Function supports UPDATE, INSERT AND DELETE
// returns int64, bool
// returns last inserted id on INSERT
// returns number of rows affected on UPDATE and DELETE

// func (db *DB) Execute(query string, args ...interface{}) (int64, bool) {
// 	stmt, err := db.db.Prepare(query)
// 	if err != nil {
// 		db.Err = err
// 		return 0, false
// 	}
// 	defer stmt.Close()
// 	res, err := stmt.Exec(args...)
// 	if err != nil {
// 		db.Err = err
// 		return 0, false
// 	}

// 	// execution is successful, now we need to return result
// 	if lastInsert, err := res.LastInsertId(); err == nil {
// 		return lastInsert, true
// 	}
// 	if rowsAffected, err := res.RowsAffected(); err == nil {
// 		return rowsAffected, true
// 	}
// 	return 0, true
// }

func (db *DB) ExecuteGetError(query string, args ...interface{}) (int64, bool, error) {
	stmt, err := db.db.Prepare(query)
	if err != nil {
		db.Err = err
		return 0, false, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(args...)
	if err != nil {
		db.Err = err
		return 0, false, err
	}

	// execution is successful, now we need to return result
	if lastInsert, err := res.LastInsertId(); err == nil {
		return lastInsert, true, nil
	}
	if rowsAffected, err := res.RowsAffected(); err == nil {
		return rowsAffected, true, nil
	}

	return 0, true, nil
}

func (db *DB) beginTransaction() *sql.Tx {
	tx, err := db.db.Begin()
	if err == nil {
		return tx
	}
	return nil
}

func (db *DB) commit(tx *sql.Tx) bool {
	if tx != nil {
		err := tx.Commit()
		if err != nil {
			return false
		}
	}
	return true
}

func (db *DB) rollback(tx *sql.Tx) bool {
	if tx != nil {
		err := tx.Rollback()
		if err != nil {
			return false
		}
	}
	return true
}

func generatePlaceholders(count int, fields int) string {
	if count == 0 {
		return ""
	}

	result := ""
	for i := 0; i < count; i++ {
		result += generatePlaceholder(fields) + ","
	}
	return result[0 : len(result)-1]
}

func generatePlaceholder(count int) string {
	if count == 0 {
		return ""
	}

	result := "("
	for i := 0; i < count; i++ {
		result += "?,"
	}
	result = result[0 : len(result)-1]
	result += ")"
	return result
}

func (db *DB) MultiInsertInit(query string, count int, cols []string) *MultiInsertHolder {
	multiInsertHolder := MultiInsertHolder{}

	query += " ( " + strings.Join(cols, ",") + " ) "
	query += " VALUES "
	query += generatePlaceholders(count, len(cols))

	multiInsertHolder.Query = query

	return &multiInsertHolder
}

func (db *DB) MultiInsertPush(holder *MultiInsertHolder, args ...interface{}) {
	holder.Data = append(holder.Data, args...)
}

func (db *DB) MultiInsertCommit(holder *MultiInsertHolder) bool {
	tx := db.beginTransaction()
	// return false
	stmt, err1 := tx.Prepare(holder.Query)

	if err1 != nil {
		db.Err = err1
		return false
	}

	defer stmt.Close()
	_, err2 := stmt.Exec(holder.Data...)
	if err2 != nil {
		db.Err = err2
		return false
	}

	return db.commit(tx)
}

func (db *DB) MultiInsert(query string, cols []string, count int, args ...interface{}) bool {
	holder := db.MultiInsertInit(query, count, cols)
	db.MultiInsertPush(holder, args...)
	return db.MultiInsertCommit(holder)
}

func (db *DB) MultiUpsert(query string, cols []string, count int, duplicateCols []string, args ...interface{}) bool {
	holder := db.MultiUpsertInit(query, count, cols, duplicateCols)
	db.MultiInsertPush(holder, args...)
	return db.MultiInsertCommit(holder)
}

func (db *DB) MultiUpsertInit(query string, count int, cols []string, duplicateCols []string) *MultiInsertHolder {
	multiUpsertHolder := MultiInsertHolder{}

	query += " ( " + strings.Join(cols, ",") + " ) "
	query += " VALUES "
	query += generatePlaceholders(count, len(cols))
	query += " ON DUPLICATE KEY UPDATE "
	upsertValues := []string{}
	for _, col := range duplicateCols {
		upsertValues = append(upsertValues, col+" = VALUES("+col+")")
	}
	query += strings.Join(upsertValues, ",")
	multiUpsertHolder.Query = query

	return &multiUpsertHolder
}

func (db *DB) InsertBulk(query string, data []map[string]string) int {
	fails := 0
	chunkSize := 1000

	totalData := len(data)
	totalChunks := int(math.Ceil(float64(totalData) / float64(chunkSize)))

	// getting the keys from the first index only
	dataKeys := []string{}

	for k := range data[0] {
		dataKeys = append(dataKeys, k)
	}

	for i := 0; i < totalChunks; i++ {

		// chunked
		upperBound := (i + 1) * chunkSize

		if upperBound > len(data) {
			upperBound = len(data)
		}

		partData := data[i*chunkSize : upperBound]

		if len(partData) == 0 {
			continue
		}

		holder := db.MultiInsertInit(query, len(partData), dataKeys)

		for _, item := range partData {

			flatData := make([]interface{}, 0, len(item))
			for _, key := range dataKeys {
				flatData = append(flatData, item[key])
			}

			db.MultiInsertPush(holder, flatData...)

		}

		ok := db.MultiInsertCommit(holder)

		if !ok {
			fails += (upperBound - i*chunkSize)

			log.Printf("Query failed = %v , upperBound = %v, len(partData) = %v, fails = %v \n", query, upperBound, len(partData), fails)

			log.Println("🚨 Here's the db error for the failed insertions.")

			log.Println("🚨 START 🚨")

			log.Println(db.Err)

			log.Println("🚨 END 🚨")
		}

	}

	return fails
}
