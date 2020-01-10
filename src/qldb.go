package main

// github.com/mattn/go-sqlite3

// This project is using the following golang internal database:
// https://godoc.org/modernc.org/ql

// go build modernc.org/ql/ql
// go install modernc.org/ql/ql

// https://stackoverflow.com/questions/21986780/is-it-possible-to-retrieve-a-column-value-by-name-using-golang-database-sql

// https://stackoverflow.com/questions/21986780/is-it-possible-to-retrieve-a-column-value-by-name-using-golang-database-sql

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/sqlite3dump"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	knownApps []string
)

type dbcon struct {
	db        *sql.DB
	masterKey []byte
	hash      []byte
}

func dbExists(filepath *string) bool {
	dbfile := "./databunker.db"
	if filepath != nil {
		if len(*filepath) > 0 {
			dbfile = *filepath
		}
	}
	if _, err := os.Stat(dbfile); os.IsNotExist(err) {
		return false
	}
	return true
}

func newDB(masterKey []byte, filepath *string) (dbcon, error) {
	dbobj := dbcon{nil, nil, nil}
	dbfile := "./databunker.db"
	if filepath != nil {
		if len(*filepath) > 0 {
			dbfile = *filepath
		}
	}
	fmt.Printf("Databunker db file is: %s\n", dbfile)
	// collect list of all tables
	/*
		if _, err := os.Stat(dbfile); !os.IsNotExist(err) {
			db2, err := ql.OpenFile(dbfile, &ql.Options{FileFormat: 2})
			if err != nil {
				return dbobj, err
			}
			dbinfo, err := db2.Info()
			for _, v := range dbinfo.Tables {
				knownApps = append(knownApps, v.Name)
			}
			db2.Close()
		}
	*/

	//ql.RegisterDriver2()
	//db, err := sql.Open("ql2", dbfile)
	db, err := sql.Open("sqlite3", "file:"+dbfile+"?_journal_mode=WAL")
	if err != nil {
		log.Fatalf("Failed to open databunker.db file: %s", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
	}
	_, err = db.Exec("vacuum")
	if err != nil {
		log.Fatalf("Error on vacuum database command")
	}
	hash := md5.Sum(masterKey)
	dbobj = dbcon{db, masterKey, hash[:]}

	// load all table names
	q := "select name from sqlite_master where type ='table'"
	tx, err := dbobj.db.Begin()
	if err != nil {
		return dbobj, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(q)
	for rows.Next() {
		t := ""
		rows.Scan(&t)
		knownApps = append(knownApps, t)
	}
	tx.Commit()
	fmt.Printf("tables: %s\n", knownApps)
	return dbobj, nil
}

func (dbobj dbcon) initDB() error {
	var err error
	if err = initUsers(dbobj.db); err != nil {
		return err
	}
	if err = initXTokens(dbobj.db); err != nil {
		return err
	}
	if err = initAudit(dbobj.db); err != nil {
		return err
	}
	if err = initConsent(dbobj.db); err != nil {
		return err
	}
	if err = initSessions(dbobj.db); err != nil {
		return err
	}
	if err = initRequests(dbobj.db); err != nil {
		return err
	}
	if err = initSharedRecords(dbobj.db); err != nil {
		return err
	}
	return nil
}

func (dbobj dbcon) closeDB() {
	dbobj.db.Close()
}

func (dbobj dbcon) backupDB(w http.ResponseWriter) {
	err := sqlite3dump.DumpDB(dbobj.db, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("error in backup: %s", err)
	}
}

func (dbobj dbcon) initUserApps() error {
	return nil
}

func escapeName(name string) string {
	if name == "when" {
		name = "`when`"
	}
	return name
}

func decodeFieldsValues(data interface{}) (string, []interface{}) {
	fields := ""
	values := make([]interface{}, 0)

	switch t := data.(type) {
	case primitive.M:
		//fmt.Println("decodeFieldsValues format is: primitive.M")
		for idx, val := range data.(primitive.M) {
			if len(fields) == 0 {
				fields = escapeName(idx)
			} else {
				fields = fields + "," + escapeName(idx)
			}
			values = append(values, val)
		}
	case *primitive.M:
		//fmt.Println("decodeFieldsValues format is: *primitive.M")
		for idx, val := range *data.(*primitive.M) {
			if len(fields) == 0 {
				fields = escapeName(idx)
			} else {
				fields = fields + "," + escapeName(idx)
			}
			values = append(values, val)
		}
	case map[string]interface{}:
		//fmt.Println("decodeFieldsValues format is: map[string]interface{}")
		for idx, val := range data.(map[string]interface{}) {
			if len(fields) == 0 {
				fields = escapeName(idx)
			} else {
				fields = fields + "," + escapeName(idx)
			}
			values = append(values, val)
		}
	default:
		fmt.Printf("XXXXXX wrong type: %T\n", t)
	}
	return fields, values
}

func decodeForCleanup(data interface{}) string {
	fields := ""

	switch t := data.(type) {
	case primitive.M:
		for idx := range data.(primitive.M) {
			if len(fields) == 0 {
				fields = escapeName(idx) + "=null"
			} else {
				fields = fields + "," + escapeName(idx) + "=null"
			}
		}
		return fields
	case map[string]interface{}:
		for idx := range data.(map[string]interface{}) {
			if len(fields) == 0 {
				fields = escapeName(idx) + "=null"
			} else {
				fields = fields + "," + escapeName(idx) + "=null"
			}
		}
	default:
		fmt.Printf("decodeForCleanup: wrong type: %s\n", t)
	}

	return fields
}

func decodeForUpdate(bdoc *bson.M, bdel *bson.M) (string, []interface{}) {
	values := make([]interface{}, 0)
	fields := ""

	if bdoc != nil {
		/*
			switch t := *bdoc.(type) {
			default:
				fmt.Printf("Type is %T\n", t)
			}
		*/

		for idx, val := range *bdoc {
			values = append(values, val)
			if len(fields) == 0 {
				fields = escapeName(idx) + "=$1"
			} else {
				fields = fields + "," + escapeName(idx) + "=$" + (strconv.Itoa(len(values)))
			}
		}
	}

	if bdel != nil {
		for idx := range *bdel {
			if len(fields) == 0 {
				fields = escapeName(idx) + "=null"
			} else {
				fields = fields + "," + escapeName(idx) + "=null"
			}
		}
	}
	return fields, values
}

func getTable(t Tbl) string {
	switch t {
	case TblName.Users:
		return "users"
	case TblName.Audit:
		return "audit"
	case TblName.Consent:
		return "consent"
	case TblName.Xtokens:
		return "xtokens"
	case TblName.Sessions:
		return "sessions"
	case TblName.Requests:
		return "requests"
	case TblName.Sharedrecords:
		return "sharedrecords"
	}
	return "users"
}

func (dbobj dbcon) createRecordInTable(tbl string, data interface{}) (int, error) {
	fields, values := decodeFieldsValues(data)
	valuesInQ := "$1"
	for idx := range values {
		if idx > 0 {
			valuesInQ = valuesInQ + ",$" + (strconv.Itoa(idx + 1))
		}
	}
	q := "insert into " + tbl + " (" + fields + ") values (" + valuesInQ + ")"
	//fmt.Printf("values: %s\n", values...)
	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	_, err = tx.Exec(q, values...)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return 1, nil
}

func (dbobj dbcon) createRecord(t Tbl, data interface{}) (int, error) {
	//if reflect.TypeOf(value) == reflect.TypeOf("string")
	tbl := getTable(t)
	return dbobj.createRecordInTable(tbl, data)
}

func (dbobj dbcon) countRecords(t Tbl, keyName string, keyValue string) (int64, error) {
	tbl := getTable(t)
	q := "select count(*) from " + tbl + " WHERE " + escapeName(keyName) + "=$1"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	row := tx.QueryRow(q, keyValue)
	// Columns
	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return int64(count), nil
}

func (dbobj dbcon) countRecords2(t Tbl, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	tbl := getTable(t)
	q := "select count(*) from " + tbl + " WHERE " + escapeName(keyName) + "=$1" +
		" AND " + escapeName(keyName2) + "=$2"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	row := tx.QueryRow(q, keyValue, keyValue2)
	// Columns
	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return int64(count), nil
}
func (dbobj dbcon) updateRecord(t Tbl, keyName string, keyValue string, bdoc *bson.M) (int64, error) {
	table := getTable(t)
	filter := escapeName(keyName) + "=\"" + keyValue + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, nil)
}

func (dbobj dbcon) updateRecordInTable(table string, keyName string, keyValue string, bdoc *bson.M) (int64, error) {
	filter := escapeName(keyName) + "=\"" + keyValue + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, nil)
}

func (dbobj dbcon) updateRecord2(t Tbl, keyName string, keyValue string,
	keyName2 string, keyValue2 string, bdoc *bson.M, bdel *bson.M) (int64, error) {
	table := getTable(t)
	filter := escapeName(keyName) + "=\"" + keyValue + "\" AND " +
		escapeName(keyName2) + "=\"" + keyValue2 + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, bdel)
}

func (dbobj dbcon) updateRecordInTable2(table string, keyName string,
	keyValue string, keyName2 string, keyValue2 string, bdoc *bson.M, bdel *bson.M) (int64, error) {
	filter := escapeName(keyName) + "=\"" + keyValue + "\" AND " +
		escapeName(keyName2) + "=\"" + keyValue2 + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, bdel)
}

func (dbobj dbcon) updateRecordInTableDo(table string, filter string, bdoc *bson.M, bdel *bson.M) (int64, error) {
	op, values := decodeForUpdate(bdoc, bdel)
	q := "update " + table + " SET " + op + " WHERE " + filter
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(q, values...)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	return num, err
}

func (dbobj dbcon) getRecord(t Tbl, keyName string, keyValue string) (bson.M, error) {
	table := getTable(t)
	q := "select * from " + table + " WHERE " + escapeName(keyName) + "=$1"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getRecordInTableDo(q, values)
}

func (dbobj dbcon) getRecordInTable(table string, keyName string, keyValue string) (bson.M, error) {
	q := "select * from " + table + " WHERE " + escapeName(keyName) + "=$1"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getRecordInTableDo(q, values)
}

func (dbobj dbcon) getRecord2(t Tbl, keyName string, keyValue string,
	keyName2 string, keyValue2 string) (bson.M, error) {
	table := getTable(t)
	q := "select * from " + table + " WHERE " + escapeName(keyName) + "=$1 AND " +
		escapeName(keyName2) + "=$2"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	values = append(values, keyValue2)
	return dbobj.getRecordInTableDo(q, values)
}

func (dbobj dbcon) getRecordInTable2(table string, keyName string, keyValue string,
	keyName2 string, keyValue2 string) (bson.M, error) {
	q := "select * from " + table + " WHERE " + escapeName(keyName) + "=$1 AND " +
		escapeName(keyName2) + "=$2"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	values = append(values, keyValue2)
	return dbobj.getRecordInTableDo(q, values)
}

func (dbobj dbcon) getRecordInTableDo(q string, values []interface{}) (bson.M, error) {
	fmt.Printf("query: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(q, values...)
	if err == sql.ErrNoRows {
		fmt.Println("nothing found")
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	//fmt.Printf("names: %s\n", columnNames)
	if err := rows.Err(); err != nil {
		return nil, err
	}
	//pointers := make([]interface{}, len(columnNames))
	recBson := bson.M{}
	rows.Next()
	columnPointers := make([]interface{}, len(columnNames))
	//for i, _ := range columnNames {
	//		columnPointers[i] = new(interface{})
	//}
	columns := make([]interface{}, len(columnNames))
	for idx := range columns {
		columnPointers[idx] = &columns[idx]
	}
	err = rows.Scan(columnPointers...)
	if err == sql.ErrNoRows {
		fmt.Println("nothing found")
		return nil, nil
	}
	if err != nil {
		fmt.Printf("nothing found: %s\n", err)
		return nil, nil
	}
	for i, colName := range columnNames {
		switch t := columns[i].(type) {
		case string:
			recBson[colName] = columns[i]
		case []uint8:
			recBson[colName] = string(columns[i].([]uint8))
		case int64:
			recBson[colName] = int32(columns[i].(int64))
		case int32:
			recBson[colName] = int32(columns[i].(int32))
		case nil:
			//fmt.Printf("is nil, not interesting\n")
		default:
			fmt.Printf("field: %s - %s, unknown: %s - %T\n", colName, columns[i], t, t)
		}
	}
	err = rows.Close()
	if err == sql.ErrNoRows {
		fmt.Println("nothing found2")
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if len(recBson) == 0 {
		fmt.Println("no result!!!")
		return nil, nil
	}
	tx.Commit()
	return recBson, nil
}

func (dbobj dbcon) deleteRecord(t Tbl, keyName string, keyValue string) (int64, error) {
	tbl := getTable(t)
	return dbobj.deleteRecordInTable(tbl, keyName, keyValue)
}

func (dbobj dbcon) deleteRecordInTable(table string, keyName string, keyValue string) (int64, error) {
	q := "delete from " + table + " WHERE " + escapeName(keyName) + "=$1"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(q, keyValue)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	return num, err
}

func (dbobj dbcon) deleteRecord2(t Tbl, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	tbl := getTable(t)
	return dbobj.deleteRecordInTable2(tbl, keyName, keyValue, keyName2, keyValue2)
}

func (dbobj dbcon) deleteRecordInTable2(table string, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	q := "delete from " + table + " WHERE " + escapeName(keyName) + "=$1 AND " +
		escapeName(keyName2) + "=$2"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(q, keyValue, keyValue2)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	return num, err
}

func (dbobj dbcon) deleteDuplicate2(t Tbl, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	tbl := getTable(t)
	return dbobj.deleteDuplicateInTable2(tbl, keyName, keyValue, keyName2, keyValue2)
}

func (dbobj dbcon) deleteDuplicateInTable2(table string, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	q := "delete from " + table + " where " + escapeName(keyName) + "=$1 AND " +
		escapeName(keyName2) + "=$2 AND rowid not in " +
		"(select min(rowid) from " + table + " where " + escapeName(keyName) + "=$3 AND " +
		escapeName(keyName2) + "=$4)"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(q, keyValue, keyValue2, keyValue, keyValue2)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	return num, err
}

func (dbobj dbcon) deleteExpired0(t Tbl, expt int32) (int64, error) {
	table := getTable(t)
	now := int32(time.Now().Unix())
	q := fmt.Sprintf("delete from %s WHERE `when`>0 AND `when`<%d", table, now-expt)
	fmt.Printf("q: %s\n", q)
	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(q)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	// vacuum database
	dbobj.db.Exec("vacuum")
	return num, err
}

func (dbobj dbcon) deleteExpired(t Tbl, keyName string, keyValue string) (int64, error) {
	table := getTable(t)
	q := "delete from " + table + " WHERE endtime>0 AND endtime<$1 AND " + escapeName(keyName) + "=$2"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	now := int32(time.Now().Unix())
	result, err := tx.Exec(q, now, keyValue)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	return num, err
}

func (dbobj dbcon) cleanupRecord(t Tbl, keyName string, keyValue string, data interface{}) (int64, error) {
	tbl := getTable(t)
	cleanup := decodeForCleanup(data)
	q := "update " + tbl + " SET " + cleanup + " WHERE " + escapeName(keyName) + "=$1"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(q, keyValue)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	return num, err
}

func (dbobj dbcon) getExpiring(t Tbl, keyName string, keyValue string) ([]bson.M, error) {
	table := getTable(t)
	now := int32(time.Now().Unix())
	q := fmt.Sprintf("select * from %s WHERE endtime>0 AND endtime<%d AND %s=$1", table, now, escapeName(keyName))
	fmt.Printf("q: %s\n", q)
	return dbobj.getListDo(q, keyValue)
}

func (dbobj dbcon) getList(t Tbl, keyName string, keyValue string, start int32, limit int32) ([]bson.M, error) {
	table := getTable(t)
	if limit > 100 {
		limit = 100
	}

	q := "select * from " + table + " WHERE " + escapeName(keyName) + "=$1"
	if start > 0 {
		q = q + " LIMIT " + strconv.FormatInt(int64(limit), 10) +
			" OFFSET " + strconv.FormatInt(int64(start), 10)
	} else if limit > 0 {
		q = q + " LIMIT " + strconv.FormatInt(int64(limit), 10)
	}
	fmt.Printf("q: %s\n", q)
	return dbobj.getListDo(q, keyValue)
}

func (dbobj dbcon) getListDo(q string, keyValue string) ([]bson.M, error) {
	tx, err := dbobj.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(q, keyValue)
	if err == sql.ErrNoRows {
		fmt.Println("nothing found")
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	//fmt.Printf("names: %s\n", columnNames)
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	var results []bson.M
	//pointers := make([]interface{}, len(columnNames))
	//rows.Next()
	for rows.Next() {
		recBson := bson.M{}
		//fmt.Println("parsing result line")
		columnPointers := make([]interface{}, len(columnNames))
		//for i, _ := range columnNames {
		//		columnPointers[i] = new(interface{})
		//}
		columns := make([]interface{}, len(columnNames))
		for idx := range columns {
			columnPointers[idx] = &columns[idx]
		}

		err = rows.Scan(columnPointers...)
		if err == sql.ErrNoRows {
			fmt.Println("nothing found")
			return nil, nil
		}
		if err != nil {
			fmt.Printf("nothing found: %s\n", err)
			return nil, nil
		}
		for i, colName := range columnNames {
			switch t := columns[i].(type) {
			case string:
				recBson[colName] = columns[i]
			case []uint8:
				recBson[colName] = string(columns[i].([]uint8))
			case int64:
				recBson[colName] = int32(columns[i].(int64))
			case nil:
				//fmt.Printf("is nil, not interesting\n")
			default:
				fmt.Printf("field: %s - %s, unknown: %s - %T\n", colName, columns[i], t, t)
			}
		}
		results = append(results, recBson)
	}
	err = rows.Close()
	if err == sql.ErrNoRows {
		fmt.Println("nothing found2")
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		fmt.Println("no result!!!")
		return nil, nil
	}
	tx.Commit()
	return results, nil
}

func (dbobj dbcon) getAllTables() ([]string, error) {
	return knownApps, nil
}

func (dbobj dbcon) validateNewApp(appName string) bool {
	if contains(knownApps, appName) == true {
		return true
	}
	if len(knownApps) >= 10 {
		return false
	}
	return true
}

func (dbobj dbcon) indexNewApp(appName string) {
	if contains(knownApps, appName) == false {
		// it is a new app, create an index
		fmt.Printf("This is a new app, creating index for :%s\n", appName)

		tx, err := dbobj.db.Begin()
		if err != nil {
			return
		}
		defer tx.Rollback()
		_, err = tx.Exec("CREATE TABLE IF NOT EXISTS " + appName + ` (
	  		token STRING,
			md5 STRING,
			rofields STRING,
	  		data STRING,
	  		status STRING,
	  		` + "`when` int);")
		if err != nil {
			return
		}
		_, err = tx.Exec("CREATE INDEX IF NOT EXISTS " + appName + "_token ON " + appName + " (token)")
		if err != nil {
			return
		}
		if err = tx.Commit(); err != nil {
			return
		}
		knownApps = append(knownApps, appName)
	}
	return
}

func initUsers(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS users (
	  token STRING,
	  key STRING,
	  md5 STRING,
	  loginidx STRING,
	  emailidx STRING,
	  phoneidx STRING,
	  rofields STRING,
	  tempcodeexp int,
	  tempcode int,
	  data STRING
	);
	`)
	if err != nil {
		return err
	}
	//fmt.Println("going to create indexes")
	_, err = tx.Exec(`CREATE INDEX users_token ON users (token);`)
	if err != nil {
		//fmt.Println("error in create index")
		return err
	}
	_, err = tx.Exec(`CREATE INDEX users_login ON users (loginidx);`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX users_email ON users (emailidx);`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX users_phone ON users (phoneidx);`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func initXTokens(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS xtokens (
	  xtoken STRING,
	  token STRING,
	  type STRING,
	  app STRING,
	  fields STRING,
	  endtime int
	);
	`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX xtokens_xtoken ON xtokens (xtoken);`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX xtokens_type ON xtokens (type);`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func initSharedRecords(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS sharedrecords (
	  token STRING,
	  record STRING,
	  partner STRING,
	  sesion STRING,
	  app STRING,
	  fields STRING,
	  endtime int,
	  ` + "`when` int);")
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX sharedrecords_record ON sharedrecords (record);`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func initAudit(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS audit (
	  atoken STRING,
	  identity STRING,
	  record STRING,
	  who STRING,
	  mode STRING,
	  app STRING,
	  title STRING,
	  status STRING,
	  msg STRING,
	  debug STRING,
	  before STRING,
	  after STRING,
	  ` + "`when` int);")
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX audit_atoken ON audit (atoken);`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX audit_record ON audit (record);`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func initRequests(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS requests (
	  rtoken STRING,
	  identity STRING,
	  token STRING,
	  mode STRING,
	  app STRING,
	  action STRING,
	  status STRING,
	  change STRING,
	  creationtime int,
	  ` + "`when` int);")
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX requests_rtoken ON requests (rtoken);`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX requests_token ON requests (token);`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func initConsent(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS consent (
	  who STRING,
	  mode STRING,
	  token STRING,
	  brief STRING,
	  status STRING,
	  message STRING,
	  freetext STRING,
	  lawfulbasis STRING,
	  consentmethod STRING,
	  referencecode STRING,
	  lastmodifiedby STRING,
	  creationtime int,
	  starttime int,
	  endtime int,
	  ` + "`when` int);")
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX consent_token ON consent (token);`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func initSessions(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS sessions (
	  token STRING,
	  session STRING,
	  data STRING,
	  endtime int,
	  ` + "`when` int);")
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX sessions_token ON sessions (token);`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX sessions_session ON sessions (session);`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
