package storage

import (
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"net/url"
	"os"
)

// Tbl is used to store table id
type Tbl int

// listTbls used to store list of tables
type listTbls struct {
	Users                Tbl
	Audit                Tbl
	Xtokens              Tbl
	Sessions             Tbl
	Requests             Tbl
	Userapps             Tbl
	Legalbasis           Tbl
	Agreements           Tbl
	Sharedrecords        Tbl
	Processingactivities Tbl
}

// TblName is enum of tables
var TblName = &listTbls{
	Users:                0,
	Audit:                1,
	Xtokens:              2,
	Sessions:             3,
	Requests:             4,
	Userapps:             5,
	Legalbasis:           6,
	Agreements:           7,
	Sharedrecords:        8,
	Processingactivities: 9,
}

var (
        allTables []string
)

func GetTable(t Tbl) string {
	switch t {
	case TblName.Users:
		return "users"
	case TblName.Audit:
		return "audit"
	case TblName.Xtokens:
		return "xtokens"
	case TblName.Sessions:
		return "sessions"
	case TblName.Requests:
		return "requests"
	case TblName.Userapps:
		return "userapps"
	case TblName.Legalbasis:
		return "legalbasis"
	case TblName.Agreements:
		return "agreements"
	case TblName.Sharedrecords:
		return "sharedrecords"
	case TblName.Processingactivities:
		return "processingactivities"
	}
	return "users"
}

type BackendDB interface {
	DBExists(*string) bool
	OpenDB(*string) error
	InitDB(*string) error
	CreateTestDB() string
	Ping() error
	CloseDB()
	BackupDB(http.ResponseWriter)
	CreateNewAppTable(string)
	Exec(string) error
	CreateRecordInTable(string, interface{}) (int, error)
	CreateRecord(Tbl, interface{}) (int, error)
	CountRecords0(Tbl) (int64, error)
	CountRecords(Tbl, string, string) (int64, error)
	UpdateRecord(Tbl, string, string, *bson.M) (int64, error)
	UpdateRecordInTable(string, string, string, *bson.M) (int64, error)
	UpdateRecord2(Tbl, string, string, string, string, *bson.M, *bson.M) (int64, error)
	UpdateRecordInTable2(string, string, string, string, string, *bson.M, *bson.M) (int64, error)
	LookupRecord(Tbl, bson.M) (bson.M, error)
	GetRecord(Tbl, string, string) (bson.M, error)
	GetRecordFromTable(string, string, string) (bson.M, error)
	GetRecord2(Tbl, string, string, string, string) (bson.M, error)
	DeleteRecord(Tbl, string, string) (int64, error)
	DeleteRecordInTable(string, string, string) (int64, error)
	DeleteRecord2(Tbl, string, string, string, string) (int64, error)
	DeleteExpired0(Tbl, int32) (int64, error)
	DeleteExpired(Tbl, string, string) (int64, error)
	CleanupRecord(Tbl, string, string, interface{}) (int64, error)
	GetExpiring(Tbl, string, string) ([]bson.M, error)
	GetUniqueList(Tbl, string) ([]bson.M, error)
	GetList0(Tbl, int32, int32, string) ([]bson.M, error)
	GetList(Tbl, string, string, int32, int32, string) ([]bson.M, error)
	GetAllTables() ([]string, error)
	ValidateNewApp(appName string) bool
}

func getDBObj() BackendDB {
	var db BackendDB
	databaseURL := os.Getenv("DATABASE_URL")
	// Check if DATABASE_URL is set and is a PostgreSQL URL
	if len(databaseURL) > 0 {
		u, err := url.Parse(databaseURL)
		if err == nil && u.Scheme == "postgres" {
			db = &PGSQLDB{}
			return db
		}
	}

	host := os.Getenv("MYSQL_HOST")
	if len(host) > 0 {
		db = &MySQLDB{}
		return db
	}
	host = os.Getenv("PGSQL_HOST")
	if len(host) > 0 {
		db = &PGSQLDB{}
	} else {
		db = &SQLiteDB{}
	}
	return db
}

// InitDB function creates tables and indexes
func InitDB(dbname *string) (BackendDB, error) {
	db := getDBObj()
	err := db.InitDB(dbname)
	return db, err
}

func OpenDB(dbname *string) (BackendDB, error) {
	db := getDBObj()
	err := db.OpenDB(dbname)
	return db, err
}

func DBExists(filepath *string) bool {
	db := getDBObj()
	return db.DBExists(filepath)
}

func CreateTestDB() string {
	db := getDBObj()
	return db.CreateTestDB()
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}

func isContainer() bool {
	//if _, err := os.Stat("/.dockerenv"); err == nil {
	//	return true
	//}
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) > 0 {
		return true
	}
	if _, err := os.Stat("/var/run/secrets/kubernetes.io"); err == nil {
		return true
	}
	return false
}
