package databunker

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbobj dbcon) getUserApp(userTOKEN string, appName string) ([]byte, error) {

	record, err := dbobj.getRecordInTable("app_"+appName, "token", userTOKEN)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	encData0 := record["data"].(string)
	return dbobj.userDecrypt(userTOKEN, encData0)
}

func (dbobj dbcon) createAppRecord(jsonData []byte, userTOKEN string, appName string, event *auditEvent) (string, error) {
	fmt.Printf("createAppRecord app is : %s\n", appName)
	encodedStr, err := dbobj.userEncrypt(userTOKEN, jsonData)
	if err != nil {
		return userTOKEN, err
	}
	dbobj.indexNewApp("app_" + appName)

	//var bdoc interface{}
	bdoc := bson.M{}
	bdoc["data"] = encodedStr
	//it is ok to use md5 here, it is only for data sanity
	md5Hash := md5.Sum([]byte(encodedStr))
	bdoc["md5"] = base64.StdEncoding.EncodeToString(md5Hash[:])
	bdoc["token"] = userTOKEN
	if event != nil {
		event.After = encodedStr
		event.App = appName
		event.Record = userTOKEN
	}
	//fmt.Println("creating new app")
	record, err := dbobj.getRecordInTable("app_"+appName, "token", userTOKEN)
	if err != nil {
		return userTOKEN, err
	}
	if record != nil {
		fmt.Println("update user app")
		_, err = dbobj.updateRecordInTable("app_"+appName, "token", userTOKEN, &bdoc)
	} else {
		_, err = dbobj.createRecordInTable("app_"+appName, bdoc)
	}
	return userTOKEN, err
}

func (dbobj dbcon) updateAppRecord(jsonDataPatch []byte, userTOKEN string, appName string, event *auditEvent) (string, error) {
	//_, err = collection.InsertOne(context.TODO(), bson.M{"name": "The Go Language2", "genre": "Coding", "authorId": "4"})
	userBson, err := dbobj.lookupUserRecord(userTOKEN)
	if userBson == nil || err != nil {
		// not found
		return userTOKEN, err
	}
	// get user key
	userKey := userBson["key"].(string)
	recordKey, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return userTOKEN, err
	}

	record, err := dbobj.getRecordInTable("app_"+appName, "token", userTOKEN)
	if err != nil {
		return userTOKEN, err
	}
	if record == nil {
		return userTOKEN, errors.New("user app record not found")
	}
	sig := record["md5"].(string)
	encData0 := record["data"].(string)
	encData, err := base64.StdEncoding.DecodeString(encData0)
	if err != nil {
		return userTOKEN, err
	}
	decrypted, err := decrypt(dbobj.masterKey, recordKey, encData)
	if err != nil {
		return userTOKEN, err
	}

	// merge
	fmt.Printf("old json: %s\n", decrypted)
	fmt.Printf("json patch: %s\n", jsonDataPatch)
	newJSON, err := jsonpatch.MergePatch(decrypted, jsonDataPatch)
	if err != nil {
		return userTOKEN, err
	}
	fmt.Printf("result: %s\n", newJSON)
	bdoc := bson.M{}
	encoded, err := encrypt(dbobj.masterKey, recordKey, newJSON)
	if err != nil {
		return userTOKEN, err
	}
	encodedStr := base64.StdEncoding.EncodeToString(encoded)
	bdoc["data"] = encodedStr
	//it is ok to use md5 here, it is only for data sanity
	md5Hash := md5.Sum([]byte(encodedStr))
	bdoc["md5"] = base64.StdEncoding.EncodeToString(md5Hash[:])
	bdoc["token"] = userTOKEN

	// here I add md5 of the original record to filter
	// to make sure this record was not change by other thread
	fmt.Println("update user app")
	result, err := dbobj.updateRecordInTable2("app_"+appName, "token", userTOKEN, "md5", sig, &bdoc, nil)
	if err != nil {
		return userTOKEN, err
	}
	if event != nil {
		event.Before = encData0
		event.After = encodedStr
		if result > 0 {
			event.Status = "ok"
		} else {
			event.Status = "failed"
			event.Msg = "failed to update"
		}
	}
	return userTOKEN, nil
}

// go over app collections and check if we have user record inside
func (dbobj dbcon) listUserApps(userTOKEN string) ([]byte, error) {
	//_, err = collection.InsertOne(context.TODO(), bson.M{"name": "The Go Language2", "genre": "Coding", "authorId": "4"})
	record, err := dbobj.lookupUserRecord(userTOKEN)
	if record == nil || err != nil {
		// not found
		return nil, err
	}
	allCollections, err := dbobj.getAllTables()
	if err != nil {
		return nil, err
	}
	var result []string
	for _, colName := range allCollections {
		if strings.HasPrefix(colName, "app_") {
			record, err := dbobj.getRecordInTable(colName, "token", userTOKEN)
			if err != nil {
				return nil, err
			}
			if record != nil {
				result = append(result, colName[4:])
			}
		}
	}
	fmt.Printf("returning: %s\n", result)
	resultJSON, err := json.Marshal(result)
	return resultJSON, err
}

func (dbobj dbcon) listAllAppsOnly() ([]string, error) {
	//fmt.Println("dump list of collections")
	allCollections, err := dbobj.getAllTables()
	if err != nil {
		return nil, err
	}
	var result []string
	for _, colName := range allCollections {
		if strings.HasPrefix(colName, "app_") {
			result = append(result, colName[4:])
		}
	}
	return result, nil
}

func (dbobj dbcon) listAllApps() ([]byte, error) {
	//fmt.Println("dump list of collections")
	allCollections, err := dbobj.getAllTables()
	if err != nil {
		return nil, err
	}
	var result []string
	for _, colName := range allCollections {
		if strings.HasPrefix(colName, "app_") {
			result = append(result, colName[4:])
		}
	}
	resultJSON, err := json.Marshal(result)
	//fmt.Println(resultJSON)
	return resultJSON, err
}
