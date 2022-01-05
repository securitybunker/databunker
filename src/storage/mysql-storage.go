package storage

// https://stackoverflow.com/questions/21986780/is-it-possible-to-retrieve-a-column-value-by-name-using-golang-database-sql

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	allTables []string
)

// MySQL struct is used to store database object
type MySQLDB struct {
	db *sql.DB
}

func (dbobj MySQLDB) getConnectionString(dbname *string) string {
	user := os.Getenv("MYSQL_USER_NAME")
	pass := os.Getenv("MYSQL_USER_PASS")
	host := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	if len(user) == 0 {
		user = "root"
	}
	if len(host) == 0 {
		host = "127.0.0.1"
	}
	if len(port) == 0 {
		port = "3306"
	}
	dbnameString := ""
	if dbname != nil && len(*dbname) > 0 {
		dbnameString = *dbname
	}
	//str0 := fmt.Sprintf("%s:****@tcp(%s:%s)/%s", user, host, port, dbnameString)
	//fmt.Printf("myql connection string: %s\n", str0)
	str := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, pass, host, port, dbnameString)
	return str
}

// DBExists function checks if database exists
func (dbobj MySQLDB) DBExists(dbname *string) bool {
	connectionString := dbobj.getConnectionString(dbname)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		//log.Fatalf("Failed to open database: %s", err)
		return false
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		//log.Fatalf("Error to open database connection: %s", err.Error())
		return false
	}
	dbobj2 := MySQLDB{db}
	record, err := dbobj2.GetRecord2(TblName.Xtokens, "token", "", "type", "root")
	if record == nil || err != nil {
		dbobj2.CloseDB()
		return false
	}
	dbobj2.CloseDB()
	return true
}

// CreateTestDB creates a test db
func (dbobj MySQLDB) CreateTestDB() string {
	testDBName := "databunker_test"
	connectionString := dbobj.getConnectionString(nil)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		//log.Fatalf("Failed to open database: %s", err)
		return testDBName
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		//log.Fatalf("Error to open database connection: %s", err.Error())
		return testDBName
	}
	fmt.Printf("** recreate database: %s\n", testDBName)
	_, err = db.Exec(fmt.Sprintf("drop database %s", testDBName))
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}
	_, err = db.Exec(fmt.Sprintf("create database %s", testDBName))
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}
	return testDBName
}

// OpenDB function opens the database
func (dbobj *MySQLDB) OpenDB(dbname *string) error {
	fmt.Printf("MYSQL Databunker db name is: %s\n", *dbname)
	connectionString := dbobj.getConnectionString(dbname)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatalf("Failed to open databunker db: %s", err)
		return err
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
		return err
	}
	dbobj.db = db
	// load all table names
	q := "show tables"
	tx, err := dbobj.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.Query(q)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return err
	}
	for rows.Next() {
		t := ""
		rows.Scan(&t)
		allTables = append(allTables, t)
	}
	tx.Commit()
	fmt.Printf("tables: %s\n", allTables)
	return nil
}

// InitDB function creates tables and indexes
func (dbobj *MySQLDB) InitDB(dbname *string) error {
	//fmt.Printf("MYSQL init Databunker database: %s\n", *dbname)
	connectionString := dbobj.getConnectionString(dbname)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	dbobj.db = db
	dbobj.initUsers()
	dbobj.initXTokens()
	dbobj.initAudit()
	dbobj.initSessions()
	dbobj.initUserapps()
	dbobj.initRequests()
	dbobj.initSharedRecords()
	dbobj.initProcessingactivities()
	dbobj.initLegalbasis()
	dbobj.initAgreements()
	return nil
}

func (dbobj MySQLDB) Ping() error {
	return dbobj.db.Ping()
}

// CloseDB function closes the open database
func (dbobj *MySQLDB) CloseDB() {
	if dbobj.db != nil {
		dbobj.db.Close()
	}
}

// BackupDB function backups existing databsae and prints database structure to http.ResponseWriter
func (dbobj MySQLDB) BackupDB(w http.ResponseWriter) {
	//err := sqlite3dump.DumpDB(dbobj.db, w)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	fmt.Printf("error in backup: %s", err)
	//}
}

func (dbobj MySQLDB) escapeName(name string) string {
	if name == "when" {
		name = "`when`"
	} else if name == "key" {
		name = "`key`"
	} else if name == "before" {
		name = "`before`"
	} else if name == "change" {
		name = "`change`"
	}
	return name
}

func (dbobj MySQLDB) decodeFieldsValues(data interface{}) (string, []interface{}) {
	fields := ""
	values := make([]interface{}, 0)

	switch t := data.(type) {
	case primitive.M:
		//fmt.Println("decodeFieldsValues format is: primitive.M")
		for idx, val := range data.(primitive.M) {
			if len(fields) == 0 {
				fields = dbobj.escapeName(idx)
			} else {
				fields = fields + "," + dbobj.escapeName(idx)
			}
			values = append(values, val)
		}
	case *primitive.M:
		//fmt.Println("decodeFieldsValues format is: *primitive.M")
		for idx, val := range *data.(*primitive.M) {
			if len(fields) == 0 {
				fields = dbobj.escapeName(idx)
			} else {
				fields = fields + "," + dbobj.escapeName(idx)
			}
			values = append(values, val)
		}
	case map[string]interface{}:
		//fmt.Println("decodeFieldsValues format is: map[string]interface{}")
		for idx, val := range data.(map[string]interface{}) {
			if len(fields) == 0 {
				fields = dbobj.escapeName(idx)
			} else {
				fields = fields + "," + dbobj.escapeName(idx)
			}
			values = append(values, val)
		}
	default:
		fmt.Printf("XXXXXX wrong type: %T\n", t)
	}
	return fields, values
}

func (dbobj MySQLDB) decodeForCleanup(data interface{}) string {
	fields := ""

	switch t := data.(type) {
	case primitive.M:
		for idx := range data.(primitive.M) {
			if len(fields) == 0 {
				fields = dbobj.escapeName(idx) + "=null"
			} else {
				fields = fields + "," + dbobj.escapeName(idx) + "=null"
			}
		}
		return fields
	case map[string]interface{}:
		for idx := range data.(map[string]interface{}) {
			if len(fields) == 0 {
				fields = dbobj.escapeName(idx) + "=null"
			} else {
				fields = fields + "," + dbobj.escapeName(idx) + "=null"
			}
		}
	default:
		fmt.Printf("decodeForCleanup: wrong type: %s\n", t)
	}

	return fields
}

func (dbobj MySQLDB) decodeForUpdate(bdoc *bson.M, bdel *bson.M) (string, []interface{}) {
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
				fields = dbobj.escapeName(idx) + "=?"
			} else {
				fields = fields + "," + dbobj.escapeName(idx) + "=?"
			}
		}
	}

	if bdel != nil {
		for idx := range *bdel {
			if len(fields) == 0 {
				fields = dbobj.escapeName(idx) + "=null"
			} else {
				fields = fields + "," + dbobj.escapeName(idx) + "=null"
			}
		}
	}
	return fields, values
}

func (dbobj MySQLDB) Exec(q string) error {
	_, err := dbobj.db.Exec(q)
	return err
}

// CreateRecordInTable creates new record
func (dbobj MySQLDB) CreateRecordInTable(tbl string, data interface{}) (int, error) {
	fields, values := dbobj.decodeFieldsValues(data)
	valuesInQ := "?"
	for idx := range values {
		if idx > 0 {
			valuesInQ = valuesInQ + ",?"
		}
	}
	q := "insert into " + tbl + " (" + fields + ") values (" + valuesInQ + ")"
	//fmt.Printf("q: %s\n", q)
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

// CreateRecord creates new record
func (dbobj MySQLDB) CreateRecord(t Tbl, data interface{}) (int, error) {
	//if reflect.TypeOf(value) == reflect.TypeOf("string")
	tbl := GetTable(t)
	return dbobj.CreateRecordInTable(tbl, data)
}

// CountRecords0 returns number of records in table
func (dbobj MySQLDB) CountRecords0(t Tbl) (int64, error) {
	tbl := GetTable(t)
	q := "select count(*) from " + tbl
	//fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	row := tx.QueryRow(q)
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

// CountRecords returns number of records that match filter
func (dbobj MySQLDB) CountRecords(t Tbl, keyName string, keyValue string) (int64, error) {
	tbl := GetTable(t)
	q := "select count(*) from " + tbl + " WHERE " + dbobj.escapeName(keyName) + "=?"
	//fmt.Printf("q: %s\n", q)

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

// UpdateRecord updates database record
func (dbobj MySQLDB) UpdateRecord(t Tbl, keyName string, keyValue string, bdoc *bson.M) (int64, error) {
	table := GetTable(t)
	filter := dbobj.escapeName(keyName) + "=\"" + keyValue + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, nil)
}

// UpdateRecordInTable updates database record
func (dbobj MySQLDB) UpdateRecordInTable(table string, keyName string, keyValue string, bdoc *bson.M) (int64, error) {
	filter := dbobj.escapeName(keyName) + "=\"" + keyValue + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, nil)
}

// UpdateRecord2 updates database record
func (dbobj MySQLDB) UpdateRecord2(t Tbl, keyName string, keyValue string,
	keyName2 string, keyValue2 string, bdoc *bson.M, bdel *bson.M) (int64, error) {
	table := GetTable(t)
	filter := dbobj.escapeName(keyName) + "=\"" + keyValue + "\" AND " +
		dbobj.escapeName(keyName2) + "=\"" + keyValue2 + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, bdel)
}

// UpdateRecordInTable2 updates database record
func (dbobj MySQLDB) UpdateRecordInTable2(table string, keyName string,
	keyValue string, keyName2 string, keyValue2 string, bdoc *bson.M, bdel *bson.M) (int64, error) {
	filter := dbobj.escapeName(keyName) + "=\"" + keyValue + "\" AND " +
		dbobj.escapeName(keyName2) + "=\"" + keyValue2 + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, bdel)
}

func (dbobj MySQLDB) updateRecordInTableDo(table string, filter string, bdoc *bson.M, bdel *bson.M) (int64, error) {
	op, values := dbobj.decodeForUpdate(bdoc, bdel)
	q := "update " + table + " SET " + op + " WHERE " + filter
	//fmt.Printf("q: %s\n", q)

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

// Lookup record by multiple fields
func (dbobj MySQLDB) LookupRecord(t Tbl, row bson.M) (bson.M, error) {
	table := GetTable(t)
	q := "select * from " + table + " WHERE "
	num := 1
	values := make([]interface{}, 0)
	for keyName, keyValue := range row {
		q = q + dbobj.escapeName(keyName) + "=?"
		if num < len(row) {
			q = q + " AND "
		}
		values = append(values, keyValue)
		num = num + 1
	}
	return dbobj.getRecordInTableDo(q, values)
}

// GetRecord returns specific record from database
func (dbobj MySQLDB) GetRecord(t Tbl, keyName string, keyValue string) (bson.M, error) {
	table := GetTable(t)
	q := "select * from " + table + " WHERE " + dbobj.escapeName(keyName) + "=?"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getRecordInTableDo(q, values)
}

// GetRecordInTable returns specific record from database
func (dbobj MySQLDB) GetRecordInTable(table string, keyName string, keyValue string) (bson.M, error) {
	q := "select * from " + table + " WHERE " + dbobj.escapeName(keyName) + "=?"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getRecordInTableDo(q, values)
}

// GetRecord2  returns specific record from database
func (dbobj MySQLDB) GetRecord2(t Tbl, keyName string, keyValue string,
	keyName2 string, keyValue2 string) (bson.M, error) {
	table := GetTable(t)
	q := "select * from " + table + " WHERE " + dbobj.escapeName(keyName) + "=? AND " +
		dbobj.escapeName(keyName2) + "=?"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	values = append(values, keyValue2)
	return dbobj.getRecordInTableDo(q, values)
}

func (dbobj MySQLDB) getRecordInTableDo(q string, values []interface{}) (bson.M, error) {
	//fmt.Printf("query: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(q, values...)
	if err != nil {
		if rows != nil {
			rows.Close()
		}
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	flag := rows.Next()
	if flag == false {
		//fmt.Printf("no result, flag: %t\n", flag)
		return nil, nil
	}
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
		fmt.Println("returning empty result")
		return nil, nil
	}
	if err != nil {
		if strings.Contains(err.Error(), "Rows are closed") {
			fmt.Println("returning empty result")
			return nil, nil
		}
		return nil, err
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
		case bool:
			recBson[colName] = columns[i].(bool)
		case nil:
			//fmt.Printf("is nil, not interesting\n")
		default:
			fmt.Printf("field: %s - %s, unknown: %s - %T\n", colName, columns[i], t, t)
		}
	}
	err = rows.Close()
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if len(recBson) == 0 {
		return nil, nil
	}
	tx.Commit()
	return recBson, nil
}

// DeleteRecord deletes record in database
func (dbobj MySQLDB) DeleteRecord(t Tbl, keyName string, keyValue string) (int64, error) {
	tbl := GetTable(t)
	return dbobj.DeleteRecordInTable(tbl, keyName, keyValue)
}

// DeleteRecordInTable deletes record in database
func (dbobj MySQLDB) DeleteRecordInTable(table string, keyName string, keyValue string) (int64, error) {
	q := "delete from " + table + " WHERE " + dbobj.escapeName(keyName) + "=?"
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

// DeleteRecord2 deletes record in database
func (dbobj MySQLDB) DeleteRecord2(t Tbl, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	tbl := GetTable(t)
	return dbobj.deleteRecordInTable2(tbl, keyName, keyValue, keyName2, keyValue2)
}

func (dbobj MySQLDB) deleteRecordInTable2(table string, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	q := "delete from " + table + " WHERE " + dbobj.escapeName(keyName) + "=? AND " +
		dbobj.escapeName(keyName2) + "=?"
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

/*
func (dbobj MySQLDB) deleteDuplicate2(t Tbl, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	tbl := GetTable(t)
	return dbobj.deleteDuplicateInTable2(tbl, keyName, keyValue, keyName2, keyValue2)
}
*/

/*
func (dbobj MySQLDB) deleteDuplicateInTable2(table string, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	q := "delete from " + table + " where " + dbobj.escapeName(keyName) + "=? AND " +
		escapeName(keyName2) + "=? AND rowid not in " +
		"(select min(rowid) from " + table + " where " + dbobj.escapeName(keyName) + "=? AND " +
		escapeName(keyName2) + "=?)"
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
*/

// DeleteExpired0 deletes expired records in database
func (dbobj MySQLDB) DeleteExpired0(t Tbl, expt int32) (int64, error) {
	table := GetTable(t)
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

// DeleteExpired deletes expired records in database
func (dbobj MySQLDB) DeleteExpired(t Tbl, keyName string, keyValue string) (int64, error) {
	table := GetTable(t)
	q := "delete from " + table + " WHERE endtime>0 AND endtime<? AND " + dbobj.escapeName(keyName) + "=?"
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

// CleanupRecord nullifies specific feilds in records in database
func (dbobj MySQLDB) CleanupRecord(t Tbl, keyName string, keyValue string, data interface{}) (int64, error) {
	tbl := GetTable(t)
	cleanup := dbobj.decodeForCleanup(data)
	q := "update " + tbl + " SET " + cleanup + " WHERE " + dbobj.escapeName(keyName) + "=?"
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

// GetExpiring get records that are expiring
func (dbobj MySQLDB) GetExpiring(t Tbl, keyName string, keyValue string) ([]bson.M, error) {
	table := GetTable(t)
	now := int32(time.Now().Unix())
	q := fmt.Sprintf("select * from %s WHERE endtime>0 AND endtime<%d AND %s=?",
		table, now, dbobj.escapeName(keyName))
	fmt.Printf("q: %s\n", q)
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getListDo(q, values)
}

// GetUniqueList returns a unique list of values from specific column in database
func (dbobj MySQLDB) GetUniqueList(t Tbl, keyName string) ([]bson.M, error) {
	table := GetTable(t)
	keyName = dbobj.escapeName(keyName)
	q := "select distinct " + keyName + " from " + table + " ORDER BY " + keyName
	//fmt.Printf("q: %s\n", q)
	values := make([]interface{}, 0)
	return dbobj.getListDo(q, values)
}

// GetList is used to return list of rows. It can be used to return values using pager.
func (dbobj MySQLDB) GetList0(t Tbl, start int32, limit int32, orderField string) ([]bson.M, error) {
	table := GetTable(t)
	if limit > 100 {
		limit = 100
	}

	q := "select * from " + table
	if len(orderField) > 0 {
		q = q + " ORDER BY " + dbobj.escapeName(orderField) + " DESC"
	}
	if start > 0 {
		q = q + " LIMIT " + strconv.FormatInt(int64(limit), 10) +
			" OFFSET " + strconv.FormatInt(int64(start), 10)
	} else if limit > 0 {
		q = q + " LIMIT " + strconv.FormatInt(int64(limit), 10)
	}
	//fmt.Printf("q: %s\n", q)
	values := make([]interface{}, 0)
	return dbobj.getListDo(q, values)
}

// GetList is used to return list of rows. It can be used to return values using pager.
func (dbobj MySQLDB) GetList(t Tbl, keyName string, keyValue string, start int32, limit int32, orderField string) ([]bson.M, error) {
	table := GetTable(t)
	if limit > 100 {
		limit = 100
	}
	q := "select * from " + table + " WHERE " + dbobj.escapeName(keyName) + "=?"
	if len(orderField) > 0 {
		q = q + " ORDER BY " + dbobj.escapeName(orderField) + " DESC"
	}
	if start > 0 {
		q = q + " LIMIT " + strconv.FormatInt(int64(limit), 10) +
			" OFFSET " + strconv.FormatInt(int64(start), 10)
	} else if limit > 0 {
		q = q + " LIMIT " + strconv.FormatInt(int64(limit), 10)
	}
	//fmt.Printf("q: %s\n", q)
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getListDo(q, values)
}

func (dbobj MySQLDB) getListDo(q string, values []interface{}) ([]bson.M, error) {
	tx, err := dbobj.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(q, values...)
	if err == sql.ErrNoRows {
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
	var results []bson.M
	//pointers := make([]interface{}, len(columnNames))
	//rows.Next()
	for rows.Next() {
		recBson := bson.M{}
		columnPointers := make([]interface{}, len(columnNames))
		columns := make([]interface{}, len(columnNames))
		for idx := range columns {
			columnPointers[idx] = &columns[idx]
		}

		err = rows.Scan(columnPointers...)
		if err == sql.ErrNoRows {
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
			case bool:
				recBson[colName] = columns[i].(bool)
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
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	tx.Commit()
	return results, nil
}

// GetAllTables returns all tables that exists in database
func (dbobj MySQLDB) GetAllTables() ([]string, error) {
	return allTables, nil
}

// ValidateNewApp function check if app name can be part of the table name
func (dbobj MySQLDB) ValidateNewApp(appName string) bool {
	if contains(allTables, appName) == true {
		return true
	}
	return true
}

func (dbobj MySQLDB) execQueries(queries []string) error {
	tx, err := dbobj.db.Begin()
	if err != nil {
		return err
	}
	for _, value := range queries {
		//fmt.Printf("exec: %s\n", value)
		_, err = tx.Exec(value)
		if err != nil {
			fmt.Printf("Error in q: %s\n", value)
			fmt.Printf("err: %s\n", err)
			tx.Rollback()
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// CreateNewAppTable creates a new app table and creates indexes for it.
func (dbobj MySQLDB) CreateNewAppTable(appName string) {
	if contains(allTables, appName) == false {
		// it is a new app, create an index
		log.Printf("This is a new app, creating table & index for: %s\n", appName)
		queries := []string{
			`CREATE TABLE IF NOT EXISTS ` + appName + ` (` +
				`token TINYTEXT,` +
				`md5 TINYTEXT,` +
				`rofields TINYTEXT,` +
				`data TEXT,` +
				`status TINYTEXT,` +
				"`when` int);",
			"CREATE UNIQUE INDEX " + appName + "_token ON " + appName + " (token(36));"}
		err := dbobj.execQueries(queries)
		if err == nil {
			allTables = append(allTables, appName)
		}
	}
	return
}

func (dbobj MySQLDB) initUsers() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (` +
			`token TINYTEXT,` +
			"`key` TINYTEXT," +
			`md5 TINYTEXT,` +
			`loginidx TINYTEXT,` +
			`emailidx TINYTEXT,` +
			`phoneidx TINYTEXT,` +
			`customidx TINYTEXT,` +
			`expstatus TINYTEXT,` +
			`exptoken TINYTEXT,` +
			`endtime int,` +
			`tempcodeexp int,` +
			`tempcode int,` +
			`data TEXT);`,
		`CREATE UNIQUE INDEX users_token ON users (token(36));`,
		`CREATE INDEX users_login ON users (loginidx(36));`,
		`CREATE INDEX users_email ON users (emailidx(36));`,
		`CREATE INDEX users_phone ON users (phoneidx(36));`,
		`CREATE INDEX users_custom ON users (customidx(36));`,
		`CREATE INDEX users_endtime ON users (endtime);`,
		`CREATE INDEX users_exptoken ON users (exptoken(36));`}
	return dbobj.execQueries(queries)
}

func (dbobj MySQLDB) initUserapps() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS userapps (` +
			`appname TINYTEXT,` +
			`token TINYTEXT,` +
			`md5 TINYTEXT,` +
			`data TEXT,` +
			`status TINYTEXT,` +
			"`when` int);",
		`CREATE INDEX userapps_appname ON userapps (appname(36));`,
		`CREATE UNIQUE INDEX userapps_token_appname ON userapps (token(36),appname(36));`}
	return dbobj.execQueries(queries)
}

func (dbobj MySQLDB) initXTokens() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS xtokens (` +
			`xtoken TINYTEXT,` +
			`token TINYTEXT,` +
			`type TINYTEXT,` +
			`app TINYTEXT,` +
			`fields TINYTEXT,` +
			`endtime int);`,
		`CREATE UNIQUE INDEX xtokens_xtoken ON xtokens (xtoken(36));`,
		`CREATE INDEX xtokens_type ON xtokens (type(20));`}
	return dbobj.execQueries(queries)
}

func (dbobj MySQLDB) initSharedRecords() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sharedrecords (` +
			`token TINYTEXT,` +
			`record TINYTEXT,` +
			`partner TINYTEXT,` +
			`session TINYTEXT,` +
			`app TINYTEXT,` +
			`fields TINYTEXT,` +
			`endtime int,` +
			"`when` int);",
		`CREATE INDEX sharedrecords_record ON sharedrecords (record(36));`}
	return dbobj.execQueries(queries)
}

func (dbobj MySQLDB) initAudit() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS audit (` +
			`atoken TINYTEXT,` +
			`identity TINYTEXT,` +
			`record TINYTEXT,` +
			`who TINYTEXT,` +
			`mode TINYTEXT,` +
			`app TINYTEXT,` +
			`title TINYTEXT,` +
			`status TINYTEXT,` +
			`msg TINYTEXT,` +
			`debug TINYTEXT,` +
			"`before` TEXT," +
			`after TEXT,` +
			"`when` int);",
		`CREATE INDEX audit_atoken ON audit (atoken(36));`,
		`CREATE INDEX audit_record ON audit (record(36));`}
	return dbobj.execQueries(queries)
}

func (dbobj MySQLDB) initRequests() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS requests (` +
			`rtoken TINYTEXT,` +
			`token TINYTEXT,` +
			`app TINYTEXT,` +
			`brief TINYTEXT,` +
			`action TINYTEXT,` +
			`status TINYTEXT,` +
			"`change` TINYTEXT," +
			`reason TINYTEXT,` +
			`creationtime int,` +
			"`when` int);",
		`CREATE INDEX requests_rtoken ON requests (rtoken(36));`,
		`CREATE INDEX requests_token ON requests (token(36));`,
		`CREATE INDEX requests_status ON requests (status(20));`}
	return dbobj.execQueries(queries)
}

func (dbobj MySQLDB) initProcessingactivities() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS processingactivities (` +
			`activity TINYTEXT,` +
			`title TINYTEXT,` +
			`script TEXT,` +
			`fulldesc TINYTEXT,` +
			`legalbasis TINYTEXT,` +
			`applicableto TINYTEXT,` +
			`creationtime int);`,
		`CREATE INDEX processingactivities_activity ON processingactivities (activity(36));`}
	return dbobj.execQueries(queries)
}

func (dbobj MySQLDB) initLegalbasis() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS legalbasis (` +
			`brief TINYTEXT,` +
			`status TINYTEXT,` +
			`module TINYTEXT,` +
			`shortdesc TINYTEXT,` +
			`fulldesc TEXT,` +
			`basistype TINYTEXT,` +
			`requiredmsg TINYTEXT,` +
			`usercontrol BOOLEAN,` +
			`requiredflag BOOLEAN,` +
			`creationtime int);`,
		`CREATE INDEX legalbasis_brief ON legalbasis (brief(36));`}
	return dbobj.execQueries(queries)
}

func (dbobj MySQLDB) initAgreements() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS agreements (` +
			`who TINYTEXT,` +
			`mode TINYTEXT,` +
			`token TINYTEXT,` +
			`brief TINYTEXT,` +
			`status TINYTEXT,` +
			`referencecode TINYTEXT,` +
			`lastmodifiedby TINYTEXT,` +
			`agreementmethod TINYTEXT,` +
			`creationtime int,` +
			`starttime int,` +
			`endtime int,` +
			"`when` int);",
		`CREATE INDEX agreements_token ON agreements (token(36));`,
		`CREATE INDEX agreements_brief ON agreements (brief(36));`}
	return dbobj.execQueries(queries)
}

func (dbobj MySQLDB) initSessions() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (` +
			`token TINYTEXT,` +
			`session TINYTEXT,` +
			"`key` TINYTEXT," +
			`data TEXT,` +
			`endtime int,` +
			"`when` int);",
		`CREATE INDEX sessions_a_token ON sessions (token(36));`,
		`CREATE INDEX sessions_a_session ON sessions (session(36));`}
	return dbobj.execQueries(queries)
}
