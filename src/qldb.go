package main

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
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"modernc.org/ql"
)

var (
	knownApps []string
)

type dbcon struct {
	db        *sql.DB
	masterKey []byte
	hash      []byte
}

func dbExists() bool {
	if _, err := os.Stat("databunker.db"); os.IsNotExist(err) {
		return false
	}
	return true
}

func newDB(masterKey []byte, urlurl *string) (dbcon, error) {
	dbobj := dbcon{nil, nil, nil}

	/*
		if db, err := ql.OpenFile("./bunker.db", &ql.Options{}); err != nil {
			fmt.Println("open db")
			a, err := db.Info()
			if err != nil {
				fmt.Printf("error in db info: %s\n", err)
			}
			for _, v := range a.Tables {
				fmt.Printf("dbinfo: %s\n", string(v.Name))
			}
			db.Close()
		}
	*/

	ql.RegisterDriver2()
	db, err := sql.Open("ql2", "./databunker.db")
	if err != nil {
		log.Fatalf("Failed to open databunker.db file: %s", err)
	}
	hash := md5.Sum(masterKey)
	dbobj = dbcon{db, masterKey, hash[:]}
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
	return nil
}

func (dbobj dbcon) initUserApps() error {
	return nil
}

func decodeFieldsValues(data interface{}) (string, string) {
	fields := ""
	values := ""
	str := ""

	switch t := data.(type) {
	case primitive.M:
		fmt.Println("format is: primitive.M")
		for idx, val := range data.(primitive.M) {
			if len(fields) == 0 {
				fields = idx
			} else {
				fields = fields + "," + idx
			}
			switch t := val.(type) {
			case string:
				str = "\"" + val.(string) + "\""
			case int:
				str = strconv.Itoa(val.(int))
			case int32:
				str = strconv.FormatInt(int64(val.(int32)), 10)
			default:
				fmt.Printf("wrong type: %s\n", t)
			}
			if len(values) == 0 {
				values = str
			} else {
				values = values + "," + str
			}
		}
	case *primitive.M:
		fmt.Println("format is: *primitive.M")
		for idx, val := range *data.(*primitive.M) {
			if len(fields) == 0 {
				fields = idx
			} else {
				fields = fields + "," + idx
			}
			switch t := val.(type) {
			case string:
				str = "\"" + val.(string) + "\""
			case int:
				str = strconv.Itoa(val.(int))
			case int32:
				str = strconv.FormatInt(int64(val.(int32)), 10)
			default:
				fmt.Printf("wrong type: %s\n", t)
			}
			if len(values) == 0 {
				values = str
			} else {
				values = values + "," + str
			}
		}
	case map[string]interface{}:
		fmt.Println("format is: map[string]interface{}")
		for idx, val := range data.(map[string]interface{}) {
			if len(fields) == 0 {
				fields = idx
			} else {
				fields = fields + "," + idx
			}
			switch t := val.(type) {
			case string:
				str = "\"" + val.(string) + "\""
			case int:
				str = strconv.Itoa(val.(int))
			case int32:
				str = strconv.FormatInt(int64(val.(int32)), 10)
			default:
				fmt.Printf("wrong type: %s\n", t)
			}
			if len(values) == 0 {
				values = str
			} else {
				values = values + "," + str
			}
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
		for idx, _ := range data.(primitive.M) {
			if len(fields) == 0 {
				fields = idx + "=null"
			} else {
				fields = fields + "," + idx + "=null"
			}
		}
		return fields
	case map[string]interface{}:
		for idx, _ := range data.(map[string]interface{}) {
			if len(fields) == 0 {
				fields = idx + "=null"
			} else {
				fields = fields + "," + idx + "=null"
			}
		}
	default:
		fmt.Printf("decodeForCleanup: wrong type: %s\n", t)
	}

	return fields
}

func decodeForUpdate(bdoc *bson.M, bdel *bson.M) string {
	fields := ""
	str := ""

	if bdoc != nil {
		/*
			switch t := *bdoc.(type) {
			default:
				fmt.Printf("Type is %T\n", t)
			}
		*/

		for idx, val := range *bdoc {
			switch t := val.(type) {
			case string:
				str = "\"" + val.(string) + "\""
			case int:
				str = strconv.Itoa(val.(int))
			case int32:
				str = strconv.FormatInt(int64(val.(int32)), 10)
			default:
				fmt.Printf("wrong type: %s\n", t)
			}
			if len(fields) == 0 {
				fields = idx + "=" + str
			} else {
				fields = fields + "," + idx + "=" + str
			}
		}
	}
	if bdel != nil {
		for idx, _ := range *bdel {
			if len(fields) == 0 {
				fields = idx + "=null"
			} else {
				fields = fields + "," + idx + "=null"
			}
		}
	}
	return fields
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
	}
	return "users"
}

func (dbobj dbcon) createRecordInTable(tbl string, data interface{}) (int, error) {
	fields, values := decodeFieldsValues(data)
	q := "insert into " + tbl + " (" + fields + ") values (" + values + ");"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(q)
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
	q := "select count(*) from " + tbl + " WHERE " + keyName + "'" + keyValue + "';"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	row := tx.QueryRow(q)
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
	filter := keyName + "=\"" + keyValue + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, nil)
}

func (dbobj dbcon) updateRecordInTable(table string, keyName string, keyValue string, bdoc *bson.M) (int64, error) {
	filter := keyName + "=\"" + keyValue + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, nil)
}

func (dbobj dbcon) updateRecord2(t Tbl, keyName string, keyValue string,
	keyName2 string, keyValue2 string, bdoc *bson.M, bdel *bson.M) (int64, error) {
	table := getTable(t)
	filter := keyName + "=\"" + keyValue + "\" AND " + keyName2 + "=\"" + keyValue2 + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, bdel)
}

func (dbobj dbcon) updateRecordInTable2(table string, keyName string,
	keyValue string, keyName2 string, keyValue2 string, bdoc *bson.M, bdel *bson.M) (int64, error) {
	filter := keyName + "=\"" + keyValue + "\" AND " + keyName2 + "=\"" + keyValue2 + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, bdel)
}

func (dbobj dbcon) updateRecordInTableDo(table string, filter string, bdoc *bson.M, bdel *bson.M) (int64, error) {
	op := decodeForUpdate(bdoc, bdel)
	q := "update " + table + " SET " + op + " WHERE " + filter
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
	return num, err
}

func (dbobj dbcon) getRecord(t Tbl, keyName string, keyValue string) (bson.M, error) {
	tbl := getTable(t)
	return dbobj.getRecordInTable(tbl, keyName, keyValue)
}

func (dbobj dbcon) getRecordInTable(table string, keyName string, keyValue string) (bson.M, error) {
	q := "select * from " + table + " WHERE " + keyName + "=\"" + keyValue + "\""
	return dbobj.getRecordInTableDo(q)
}

func (dbobj dbcon) getRecord2(t Tbl, keyName string, keyValue string,
	keyName2 string, keyValue2 string) (bson.M, error) {
	tbl := getTable(t)
	return dbobj.getRecordInTable2(tbl, keyName, keyValue, keyName2, keyValue2)
}

func (dbobj dbcon) getRecordInTable2(table string, keyName string, keyValue string,
	keyName2 string, keyValue2 string) (bson.M, error) {
	q := "select * from " + table + " WHERE " + keyName + "=\"" + keyValue + "\" AND " +
		keyName2 + "=\"" + keyValue2 + "\""
	return dbobj.getRecordInTableDo(q)
}

func (dbobj dbcon) getRecordInTableDo(q string) (bson.M, error) {
	fmt.Printf("q: %s\n", q)
	tx, err := dbobj.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(q)
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
	//pointers := make([]interface{}, len(columnNames))
	recBson := bson.M{}
	rows.Next()
	//for rows.Next() {
	//fmt.Println("parsing result line")
	columnPointers := make([]interface{}, len(columnNames))
	//for i, _ := range columnNames {
	//		columnPointers[i] = new(interface{})
	//}
	columns := make([]interface{}, len(columnNames))
	for i, _ := range columns {
		columnPointers[i] = &columns[i]
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
	//}
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
	if err = tx.Commit(); err != nil {
		return recBson, err
	}
	return recBson, nil
}

func (dbobj dbcon) deleteRecord(t Tbl, keyName string, keyValue string) (int64, error) {
	tbl := getTable(t)
	return dbobj.deleteRecordInTable(tbl, keyName, keyValue)
}

func (dbobj dbcon) deleteRecordInTable(table string, keyName string, keyValue string) (int64, error) {
	q := "delete from " + table + " WHERE " + keyName + "=\"" + keyValue + "\""
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
	return num, err
}

func (dbobj dbcon) cleanupRecord(t Tbl, keyName string, keyValue string, data interface{}) (int64, error) {
	tbl := getTable(t)
	cleanup := decodeForCleanup(data)
	q := "update " + tbl + " SET " + cleanup + " WHERE " + keyName + "=\"" + keyValue + "\""
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
	return num, err
}

func (dbobj dbcon) getList(t Tbl, keyName string, keyValue string, start int32, limit int32) ([]bson.M, error) {
	fmt.Println("TODO")
	return nil, nil
}

func (dbobj dbcon) getAllTables() ([]string, error) {
	//for nm, tab := range dbobj.db.root.tables {
	//
	//	}
	// HasNextResultSet()
	a := []string{"aaa"}
	a = append(a, "test123")
	return a, nil
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
	  		data STRING,
	  		status STRING,
	  		when int
		);`)
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

/*
BEGIN TRANSACTION;
	CREATE TABLE Orders (CustomerID int, Date time);
	CREATE INDEX OrdersID ON Orders (id());
	CREATE INDEX OrdersDate ON Orders (Date);
	CREATE TABLE Items (OrderID int, ProductID int, Qty int);
	CREATE INDEX ItemsOrderID ON Items (OrderID);
COMMIT;
*/

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
	  tempcode STRING,
	  tempcodeexp int,
	  data string
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
	  endtime int32
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

/*
	When   int32  `json:"when"`
	Who    string `json:"who"`
	Record string `json:"record"`
	App    string `json:"app"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Debug  string `json:"debug"`
	Before string `json:"before"`
	After  string `json:"after"`
	Meta   string `json:"meta"`
*/

func initAudit(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS audit (
	  record STRING,
	  who STRING,
	  app STRING,
	  title STRING,
	  status STRING,
	  msg STRING,
	  debug STRING,
	  before STRING,
	  after STRING,
	  meta STRING,
	  when int
	);
	`)
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

/*
	When    int32  `json:"when,omitempty"`
	Who     string `json:"who,omitempty"`
	Type    string `json:"type,omitempty"`
	Token   string `json:"token,omitempty"`
	Brief   string `json:"brief,omitempty"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
*/

func initConsent(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS consent (
	  who STRING,
	  type STRING,
	  token STRING,
	  brief STRING,
	  message STRING,
	  status STRING,
	  when int
	);
	`)
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
	  meta STRING,
	  when int,
	  endtime int
	);
	`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE INDEX sessions_token ON sessions (token);`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
