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

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PGSQL struct is used to store database object
type PGSQLDB struct {
	db *sql.DB
}

func (dbobj PGSQLDB) getConnectionString(dbname *string) string {
	user := os.Getenv("PGSQL_USER_NAME")
	pass := os.Getenv("PGSQL_USER_PASS")
	host := os.Getenv("PGSQL_HOST")
	port := os.Getenv("PGSQL_PORT")
	if len(user) == 0 {
		user = "postgres"
	}
	if len(host) == 0 {
		host = "127.0.0.1"
	}
	if len(port) == 0 {
		port = "5432"
	}
	dbnameString := ""
	if dbname != nil && len(*dbname) > 0 {
		dbnameString = *dbname
	}
	if len(os.Getenv("PGSQL_USER_PASS_FILE")) > 0 {
		content, err := os.ReadFile(os.Getenv("PGSQL_USER_PASS_FILE"))
		if err != nil {
			return ""
		}
		pass = strings.TrimSpace(string(content))
	}
	//str0 := fmt.Sprintf("%s:****@tcp(%s:%s)/%s", user, host, port, dbnameString)
	//fmt.Printf("myql connection string: %s\n", str0)
	//str := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, pass, host, port, dbnameString)
	str := fmt.Sprintf("user='%s' password='%s' host='%s' port='%s' dbname='%s'",
		user, pass, host, port, dbnameString)
	return str
}

// DBExists function checks if database exists
func (dbobj PGSQLDB) DBExists(dbname *string) bool {
	connectionString := dbobj.getConnectionString(dbname)
	db, err := sql.Open("postgres", connectionString)
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
	dbobj2 := PGSQLDB{db}
	record, err := dbobj2.GetRecord2(TblName.Xtokens, "token", "", "type", "root")
	if record == nil || err != nil {
		dbobj2.CloseDB()
		return false
	}
	dbobj2.CloseDB()
	return true
}

// CreateTestDB creates a test db
func (dbobj PGSQLDB) CreateTestDB() string {
	testDBName := "databunker_test"
	connectionString := dbobj.getConnectionString(nil)
	db, err := sql.Open("postgres", connectionString)
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
func (dbobj *PGSQLDB) OpenDB(dbname *string) error {
	fmt.Printf("POSTGRESQL Databunker db name is: %s\n", *dbname)
	connectionString := dbobj.getConnectionString(dbname)
	db, err := sql.Open("postgres", connectionString)
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
	q := "SELECT table_name FROM information_schema.tables where table_schema NOT IN ('pg_catalog', 'information_schema');"
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
func (dbobj *PGSQLDB) InitDB(dbname *string) error {
	//fmt.Printf("PGSQL init Databunker database: %s\n", *dbname)
	connectionString := dbobj.getConnectionString(dbname)
	db, err := sql.Open("postgres", connectionString)
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

func (dbobj PGSQLDB) Ping() error {
	return dbobj.db.Ping()
}

// CloseDB function closes the open database
func (dbobj *PGSQLDB) CloseDB() {
	if dbobj.db != nil {
		dbobj.db.Close()
	}
}

// BackupDB function backups existing database and prints database structure to http.ResponseWriter
func (dbobj PGSQLDB) BackupDB(w http.ResponseWriter) {
	//err := sqlite3dump.DumpDB(dbobj.db, w)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	fmt.Printf("error in backup: %s", err)
	//}
}

func (dbobj PGSQLDB) escapeName(name string) string {
	if name == "when" {
		name = `"when"`
	} else if name == "before" {
		name = `"before"`
	}
	return name
}

func (dbobj PGSQLDB) decodeFieldsValues(data interface{}) (string, []interface{}) {
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

func (dbobj PGSQLDB) decodeForCleanup(bdel []string) string {
	fields := ""
	if bdel != nil {
		for _, colname := range bdel {
			if len(fields) == 0 {
				fields = dbobj.escapeName(colname) + "=null"
			} else {
				fields = fields + "," + dbobj.escapeName(colname) + "=null"
			}
		}
	}
	return fields
}

func (dbobj PGSQLDB) decodeForUpdate(bdoc *bson.M, bdel []string) (string, []interface{}) {
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
				fields = dbobj.escapeName(idx) + "=$1"
			} else {
				fields = fields + "," + dbobj.escapeName(idx) + "=$" + (strconv.Itoa(len(values)))
			}
		}
	}

	if bdel != nil {
		for _, colname := range bdel {
			if len(fields) == 0 {
				fields = dbobj.escapeName(colname) + "=null"
			} else {
				fields = fields + "," + dbobj.escapeName(colname) + "=null"
			}
		}
	}
	return fields, values
}

func (dbobj PGSQLDB) Exec(q string) error {
	_, err := dbobj.db.Exec(q)
	return err
}

// CreateRecordInTable creates new record
func (dbobj PGSQLDB) CreateRecordInTable(tbl string, data interface{}) (int, error) {
	fields, values := dbobj.decodeFieldsValues(data)
	valuesInQ := "$1"
	for idx := range values {
		if idx > 0 {
			valuesInQ = valuesInQ + ",$" + (strconv.Itoa(idx + 1))
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
func (dbobj PGSQLDB) CreateRecord(t Tbl, data interface{}) (int, error) {
	//if reflect.TypeOf(value) == reflect.TypeOf("string")
	tbl := GetTable(t)
	return dbobj.CreateRecordInTable(tbl, data)
}

// CountRecords0 returns number of records in table
func (dbobj PGSQLDB) CountRecords0(t Tbl) (int64, error) {
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
func (dbobj PGSQLDB) CountRecords(t Tbl, keyName string, keyValue string) (int64, error) {
	tbl := GetTable(t)
	q := "select count(*) from " + tbl + " WHERE " + dbobj.escapeName(keyName) + "=$1"
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
func (dbobj PGSQLDB) UpdateRecord(t Tbl, keyName string, keyValue string, bdoc *bson.M) (int64, error) {
	table := GetTable(t)
	filter := dbobj.escapeName(keyName) + "='" + keyValue + "'"
	return dbobj.updateRecordInTableDo(table, filter, bdoc, nil)
}

// UpdateRecordInTable updates database record
func (dbobj PGSQLDB) UpdateRecordInTable(table string, keyName string, keyValue string, bdoc *bson.M) (int64, error) {
	filter := dbobj.escapeName(keyName) + "='" + keyValue + "'"
	return dbobj.updateRecordInTableDo(table, filter, bdoc, nil)
}

// UpdateRecord2 updates database record
func (dbobj PGSQLDB) UpdateRecord2(t Tbl, keyName string, keyValue string,
	keyName2 string, keyValue2 string, bdoc *bson.M, bdel []string) (int64, error) {
	table := GetTable(t)
	filter := dbobj.escapeName(keyName) + "='" + keyValue + "' AND " +
		dbobj.escapeName(keyName2) + "='" + keyValue2 + "'"
	return dbobj.updateRecordInTableDo(table, filter, bdoc, bdel)
}

// UpdateRecordInTable2 updates database record
func (dbobj PGSQLDB) UpdateRecordInTable2(table string, keyName string,
	keyValue string, keyName2 string, keyValue2 string, bdoc *bson.M, bdel []string) (int64, error) {
	filter := dbobj.escapeName(keyName) + "='" + keyValue + "' AND " +
		dbobj.escapeName(keyName2) + "='" + keyValue2 + "'"
	return dbobj.updateRecordInTableDo(table, filter, bdoc, bdel)
}

func (dbobj PGSQLDB) updateRecordInTableDo(table string, filter string, bdoc *bson.M, bdel []string) (int64, error) {
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
func (dbobj PGSQLDB) LookupRecord(t Tbl, row bson.M) (bson.M, error) {
	table := GetTable(t)
	q := "select * from " + table + " WHERE "
	num := 1
	values := make([]interface{}, 0)
	for keyName, keyValue := range row {
		q = q + dbobj.escapeName(keyName) + "=$" + strconv.FormatInt(int64(num), 10)
		if num < len(row) {
			q = q + " AND "
		}
		values = append(values, keyValue)
		num = num + 1
	}
	return dbobj.getOneRecord(q, values)
}

// GetRecord returns specific record from database
func (dbobj PGSQLDB) GetRecord(t Tbl, keyName string, keyValue string) (bson.M, error) {
	table := GetTable(t)
	q := "select * from " + table + " WHERE " + dbobj.escapeName(keyName) + "=$1"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getOneRecord(q, values)
}

// GetRecordFromTable returns specific record from database
func (dbobj PGSQLDB) GetRecordFromTable(table string, keyName string, keyValue string) (bson.M, error) {
	q := "select * from " + table + " WHERE " + dbobj.escapeName(keyName) + "=$1"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getOneRecord(q, values)
}

// GetRecord2  returns specific record from database
func (dbobj PGSQLDB) GetRecord2(t Tbl, keyName string, keyValue string,
	keyName2 string, keyValue2 string) (bson.M, error) {
	table := GetTable(t)
	q := "select * from " + table + " WHERE " + dbobj.escapeName(keyName) + "=$1 AND " +
		dbobj.escapeName(keyName2) + "=$2"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	values = append(values, keyValue2)
	return dbobj.getOneRecord(q, values)
}

func (dbobj PGSQLDB) getOneRecord(q string, values []interface{}) (bson.M, error) {
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
func (dbobj PGSQLDB) DeleteRecord(t Tbl, keyName string, keyValue string) (int64, error) {
	tbl := GetTable(t)
	return dbobj.DeleteRecordInTable(tbl, keyName, keyValue)
}

// DeleteRecordInTable deletes record in database
func (dbobj PGSQLDB) DeleteRecordInTable(table string, keyName string, keyValue string) (int64, error) {
	q := "delete from " + table + " WHERE " + dbobj.escapeName(keyName) + "=$1"
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
func (dbobj PGSQLDB) DeleteRecord2(t Tbl, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	tbl := GetTable(t)
	return dbobj.deleteRecordInTable2(tbl, keyName, keyValue, keyName2, keyValue2)
}

func (dbobj PGSQLDB) deleteRecordInTable2(table string, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	q := "delete from " + table + " WHERE " + dbobj.escapeName(keyName) + "=$1 AND " +
		dbobj.escapeName(keyName2) + "=$2"
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
func (dbobj PGSQLDB) deleteDuplicate2(t Tbl, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	tbl := GetTable(t)
	return dbobj.deleteDuplicateInTable2(tbl, keyName, keyValue, keyName2, keyValue2)
}
*/

/*
func (dbobj PGSQLDB) deleteDuplicateInTable2(table string, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	q := "delete from " + table + " where " + dbobj.escapeName(keyName) + "=$1 AND " +
		escapeName(keyName2) + "=$2 AND rowid not in " +
		"(select min(rowid) from " + table + " where " + dbobj.escapeName(keyName) + "=$3 AND " +
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
*/

// DeleteExpired0 deletes expired records in database
func (dbobj PGSQLDB) DeleteExpired0(t Tbl, expt int32) (int64, error) {
	table := GetTable(t)
	now := int32(time.Now().Unix())
	q := fmt.Sprintf(`delete from %s WHERE "when">0 AND "when"<%d`, table, now-expt)
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
func (dbobj PGSQLDB) DeleteExpired(t Tbl, keyName string, keyValue string) (int64, error) {
	table := GetTable(t)
	q := "delete from " + table + " WHERE endtime>0 AND endtime<$1 AND " + dbobj.escapeName(keyName) + "=$2"
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
func (dbobj PGSQLDB) CleanupRecord(t Tbl, keyName string, keyValue string, bdel []string) (int64, error) {
	tbl := GetTable(t)
	cleanup := dbobj.decodeForCleanup(bdel)
	q := "update " + tbl + " SET " + cleanup + " WHERE " + dbobj.escapeName(keyName) + "=$1"
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
func (dbobj PGSQLDB) GetExpiring(t Tbl, keyName string, keyValue string) ([]map[string]interface{}, error) {
	table := GetTable(t)
	now := int32(time.Now().Unix())
	q := fmt.Sprintf("select * from %s WHERE endtime>0 AND endtime<%d AND %s=$1",
		table, now, dbobj.escapeName(keyName))
	fmt.Printf("q: %s\n", q)
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getListDo(q, values)
}

// GetUniqueList returns a unique list of values from specific column in database
func (dbobj PGSQLDB) GetUniqueList(t Tbl, keyName string) ([]map[string]interface{}, error) {
	table := GetTable(t)
	keyName = dbobj.escapeName(keyName)
	q := "select distinct " + keyName + " from " + table + " ORDER BY " + keyName
	//fmt.Printf("q: %s\n", q)
	values := make([]interface{}, 0)
	return dbobj.getListDo(q, values)
}

// GetList is used to return list of rows. It can be used to return values using pager.
func (dbobj PGSQLDB) GetList0(t Tbl, start int32, limit int32, orderField string) ([]map[string]interface{}, error) {
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
func (dbobj PGSQLDB) GetList(t Tbl, keyName string, keyValue string, start int32, limit int32, orderField string) ([]map[string]interface{}, error) {
	table := GetTable(t)
	if limit > 100 {
		limit = 100
	}
	q := "select * from " + table + " WHERE " + dbobj.escapeName(keyName) + "=$1"
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

func (dbobj PGSQLDB) getListDo(q string, values []interface{}) ([]map[string]interface{}, error) {
	tx, err := dbobj.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	return dbobj.getListDoRaw(tx, q, values)
}

func (dbobj PGSQLDB) getListDoRaw(tx *sql.Tx, q string, values []interface{}) ([]map[string]interface{}, error) {
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
	var results []map[string]interface{}
	//pointers := make([]interface{}, len(columnNames))
	//rows.Next()
	for rows.Next() {
		recBson := make(map[string]interface{})
		columnPointers := make([]interface{}, len(columnNames))
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
func (dbobj PGSQLDB) GetAllTables() ([]string, error) {
	return allTables, nil
}

// ValidateNewApp function check if app name can be part of the table name
func (dbobj PGSQLDB) ValidateNewApp(appName string) bool {
	if SliceContains(allTables, appName) == true {
		return true
	}
	if len(allTables) >= 100 {
		return false
	}
	return true
}

func (dbobj PGSQLDB) execQueries(queries []string) error {
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
func (dbobj PGSQLDB) CreateNewAppTable(appName string) {
	if SliceContains(allTables, appName) == false {
		// it is a new app, create an index
		log.Printf("This is a new app, creating table & index for: %s\n", appName)
		queries := []string{
			`CREATE TABLE IF NOT EXISTS ` + appName + ` (` +
				`token VARCHAR,` +
				`md5 VARCHAR,` +
				`rofields VARCHAR,` +
				`data TEXT,` +
				`status VARCHAR,` +
				`"when" int);`,
			"CREATE UNIQUE INDEX " + appName + "_token ON " + appName + " (token);"}
		err := dbobj.execQueries(queries)
		if err == nil {
			allTables = append(allTables, appName)
		}
	}
	return
}

func (dbobj PGSQLDB) initUsers() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (` +
			`token VARCHAR,` +
			`key VARCHAR,` +
			`md5 VARCHAR,` +
			`loginidx VARCHAR,` +
			`emailidx VARCHAR,` +
			`phoneidx VARCHAR,` +
			`customidx VARCHAR,` +
			`expstatus VARCHAR,` +
			`exptoken VARCHAR,` +
			`endtime int,` +
			`tempcodeexp int,` +
			`tempcode int,` +
			`data TEXT);`,
		`CREATE UNIQUE INDEX users_token ON users (token);`,
		`CREATE INDEX users_login ON users (loginidx);`,
		`CREATE INDEX users_email ON users (emailidx);`,
		`CREATE INDEX users_phone ON users (phoneidx);`,
		`CREATE INDEX users_custom ON users (customidx);`,
		`CREATE INDEX users_endtime ON users (endtime);`,
		`CREATE INDEX users_exptoken ON users (exptoken);`}
	return dbobj.execQueries(queries)
}

func (dbobj PGSQLDB) initUserapps() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS userapps (` +
			`appname VARCHAR,` +
			`token VARCHAR,` +
			`md5 VARCHAR,` +
			`data TEXT,` +
			`status VARCHAR,` +
			`"when" int);`,
		`CREATE INDEX userapps_appname ON userapps (appname);`,
		`CREATE UNIQUE INDEX userapps_token_appname ON userapps (token,appname);`}
	return dbobj.execQueries(queries)
}

func (dbobj PGSQLDB) initXTokens() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS xtokens (` +
			`xtoken VARCHAR,` +
			`token VARCHAR,` +
			`type VARCHAR,` +
			`app VARCHAR,` +
			`fields VARCHAR,` +
			`endtime int);`,
		`CREATE UNIQUE INDEX xtokens_xtoken ON xtokens (xtoken);`,
		`CREATE INDEX xtokens_type ON xtokens (type);`}
	return dbobj.execQueries(queries)
}

func (dbobj PGSQLDB) initSharedRecords() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sharedrecords (` +
			`token VARCHAR,` +
			`record VARCHAR,` +
			`partner VARCHAR,` +
			`session VARCHAR,` +
			`app VARCHAR,` +
			`fields VARCHAR,` +
			`endtime int,` +
			`"when" int);`,
		`CREATE INDEX sharedrecords_record ON sharedrecords (record);`}
	return dbobj.execQueries(queries)
}

func (dbobj PGSQLDB) initAudit() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS audit (` +
			`atoken VARCHAR,` +
			`identity VARCHAR,` +
			`record VARCHAR,` +
			`who VARCHAR,` +
			`mode VARCHAR,` +
			`app VARCHAR,` +
			`title VARCHAR,` +
			`status VARCHAR,` +
			`msg VARCHAR,` +
			`debug VARCHAR,` +
			`before TEXT,` +
			`after TEXT,` +
			`"when" int);`,
		`CREATE INDEX audit_atoken ON audit (atoken);`,
		`CREATE INDEX audit_record ON audit (record);`}
	return dbobj.execQueries(queries)
}

func (dbobj PGSQLDB) initRequests() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS requests (` +
			`rtoken VARCHAR,` +
			`token VARCHAR,` +
			`app VARCHAR,` +
			`brief VARCHAR,` +
			`action VARCHAR,` +
			`status VARCHAR,` +
			`change VARCHAR,` +
			`reason VARCHAR,` +
			`creationtime int,` +
			`"when" int);`,
		`CREATE INDEX requests_rtoken ON requests (rtoken);`,
		`CREATE INDEX requests_token ON requests (token);`,
		`CREATE INDEX requests_status ON requests (status);`}
	return dbobj.execQueries(queries)
}

func (dbobj PGSQLDB) initProcessingactivities() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS processingactivities (` +
			`activity VARCHAR,` +
			`title VARCHAR,` +
			`script VARCHAR,` +
			`fulldesc VARCHAR,` +
			`legalbasis VARCHAR,` +
			`applicableto VARCHAR,` +
			`creationtime int);`,
		`CREATE INDEX processingactivities_activity ON processingactivities (activity);`}
	return dbobj.execQueries(queries)
}

func (dbobj PGSQLDB) initLegalbasis() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS legalbasis (` +
			`brief VARCHAR,` +
			`status VARCHAR,` +
			`module VARCHAR,` +
			`shortdesc VARCHAR,` +
			`fulldesc VARCHAR,` +
			`basistype VARCHAR,` +
			`requiredmsg VARCHAR,` +
			`usercontrol BOOLEAN,` +
			`requiredflag BOOLEAN,` +
			`creationtime int);`,
		`CREATE INDEX legalbasis_brief ON legalbasis (brief);`}
	return dbobj.execQueries(queries)
}

func (dbobj PGSQLDB) initAgreements() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS agreements (` +
			`who VARCHAR,` +
			`mode VARCHAR,` +
			`token VARCHAR,` +
			`brief VARCHAR,` +
			`status VARCHAR,` +
			`referencecode VARCHAR,` +
			`lastmodifiedby VARCHAR,` +
			`agreementmethod VARCHAR,` +
			`creationtime int,` +
			`starttime int,` +
			`endtime int,` +
			`"when" int);`,
		`CREATE INDEX agreements_token ON agreements (token);`,
		`CREATE INDEX agreements_brief ON agreements (brief);`}
	return dbobj.execQueries(queries)
}

func (dbobj PGSQLDB) initSessions() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (` +
			`token VARCHAR,` +
			`session VARCHAR,` +
			`key VARCHAR,` +
			`data TEXT,` +
			`endtime int,` +
			`"when" int);`,
		`CREATE INDEX sessions_a_token ON sessions (token);`,
		`CREATE INDEX sessions_a_session ON sessions (session);`}
	return dbobj.execQueries(queries)
}
