package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/securitybunker/databunker/src/storage"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbobj dbcon) getUserApp(userTOKEN string, appName string, conf Config) ([]byte, error) {
	appNameFull := "app_" + appName
	var record bson.M
	var err error
	if conf.Generic.UseSeparateAppTables == true {
		record, err = dbobj.store.GetRecordInTable(appNameFull, "token", userTOKEN)
	} else {
		record, err = dbobj.store.GetRecord2(storage.TblName.Userapps, "token", userTOKEN, "appname", appName)
	}
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	encData0 := record["data"].(string)
	return dbobj.userDecrypt(userTOKEN, encData0)
}

func (dbobj dbcon) deleteUserApp(userTOKEN string, appName string, conf Config) {
	appNameFull := "app_" + appName
	if conf.Generic.UseSeparateAppTables == true {
		dbobj.store.DeleteRecordInTable(appNameFull, "token", userTOKEN)
	} else {
		dbobj.store.DeleteRecord2(storage.TblName.Userapps, "token", userTOKEN, "appname", appName)
	}
}

func (dbobj dbcon) deleteUserApps(userTOKEN string, conf Config) {
	if conf.Generic.UseSeparateAppTables == true {
		userApps, _:= dbobj.listAllAppsOnly(conf)
		// delete all user app records
		for _, appName := range userApps {
			appNameFull := "app_" + appName
			dbobj.store.DeleteRecordInTable(appNameFull, "token", userTOKEN)
		}
	} else {
		dbobj.store.DeleteRecord(storage.TblName.Userapps, "token", userTOKEN)
	}
}

func (dbobj dbcon) createAppRecord(jsonData []byte, userTOKEN string, appName string, event *auditEvent, conf Config) (string, error) {
	appNameFull := "app_" + appName
	//log.Printf("Going to create app record: %s\n", appName)
	encodedStr, err := dbobj.userEncrypt(userTOKEN, jsonData)
	if err != nil {
		return userTOKEN, err
	}
	if conf.Generic.UseSeparateAppTables == true {
		dbobj.store.CreateNewAppTable(appNameFull)
	}

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
	if conf.Generic.UseSeparateAppTables == true {
		record, err := dbobj.store.GetRecordInTable(appNameFull, "token", userTOKEN)
		if err != nil {
			return userTOKEN, err
		}
		if record != nil {
			_, err = dbobj.store.UpdateRecordInTable(appNameFull, "token", userTOKEN, &bdoc)
		 } else {
			_, err = dbobj.store.CreateRecordInTable(appNameFull, bdoc)
		}
	} else {
		record, err := dbobj.store.GetRecord2(storage.TblName.Userapps, "token", userTOKEN, "appname", appName)
		if err != nil {
			return userTOKEN, err
		}
		if record != nil {
			_, err = dbobj.store.UpdateRecord2(storage.TblName.Userapps, "token", userTOKEN, "appname", appName, &bdoc, nil)
		} else {
			bdoc["appname"] = appName
			_, err = dbobj.store.CreateRecord(storage.TblName.Userapps, &bdoc)
		}
	}
	return userTOKEN, err
}

func (dbobj dbcon) updateAppRecord(jsonDataPatch []byte, userTOKEN string, appName string, event *auditEvent, conf Config) (string, error) {
	//_, err = collection.InsertOne(context.TODO(), bson.M{"name": "The Go Language2", "genre": "Coding", "authorId": "4"})
	appNameFull := "app_" + appName
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
	var record bson.M
	if conf.Generic.UseSeparateAppTables == true {
		record, err = dbobj.store.GetRecordInTable(appNameFull, "token", userTOKEN)
	} else {
		record, err = dbobj.store.GetRecord2(storage.TblName.Userapps, "token", userTOKEN, "appname", appName)
	}
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
	//fmt.Printf("old json: %s\n", decrypted)
	//fmt.Printf("json patch: %s\n", jsonDataPatch)
	var newJSON []byte
	if jsonDataPatch[0] == '{' {
		newJSON, err = jsonpatch.MergePatch(decrypted, jsonDataPatch)
	} else {
		patch, err := jsonpatch.DecodePatch(jsonDataPatch)
		if err != nil {
			return userTOKEN, err
		}
		newJSON, err = patch.Apply(decrypted)
	}
	if err != nil {
		return userTOKEN, err
	}
	//fmt.Printf("result: %s\n", newJSON)
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
	result := int64(0)
	if conf.Generic.UseSeparateAppTables == true {
		result, err = dbobj.store.UpdateRecordInTable2(appNameFull, "token", userTOKEN, "md5", sig, &bdoc, nil)
	} else {
		result, err = dbobj.store.UpdateRecord2(storage.TblName.Userapps, "token", userTOKEN, "appname", appName, &bdoc, nil)
	}
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
func (dbobj dbcon) listUserApps(userTOKEN string, conf Config) ([]byte, error) {
	//_, err = collection.InsertOne(context.TODO(), bson.M{"name": "The Go Language2", "genre": "Coding", "authorId": "4"})
	record, err := dbobj.lookupUserRecord(userTOKEN)
	if record == nil || err != nil {
		// not found
		return nil, err
	}
	var result []string
	if conf.Generic.UseSeparateAppTables == true {
		allCollections, err := dbobj.store.GetAllTables()
		if err != nil {
			return nil, err
		}
		var result []string
		for _, colName := range allCollections {
			if strings.HasPrefix(colName, "app_") {
				record, err := dbobj.store.GetRecordInTable(colName, "token", userTOKEN)
				if err != nil {
					return nil, err
				}
				if record != nil {
					result = append(result, colName[4:])
				}
			}
		}
	} else {
		records, err := dbobj.store.GetList(storage.TblName.Userapps, "token", userTOKEN, 0, 0, "appname")
		if err != nil {
			return nil, err
		}
		count := len(records)
		if count == 0 {
			return []byte("[]"), nil
		}
		for _, rec := range records {
			appname := rec["appname"].(string)
			result = append(result, appname)
		}
	}
	if len(result) == 0 {
		return []byte("[]"), nil
	}
	resultJSON, err := json.Marshal(result)
	return resultJSON, err
}

func (dbobj dbcon) dumpUserApps(userTOKEN string, conf Config) ([]byte, error) {
	results := make(map[string]interface{})
	if conf.Generic.UseSeparateAppTables == true {
		allCollections, err := dbobj.store.GetAllTables()
		if err != nil {
			return nil, err
		}
		for _, colName := range allCollections {
			if strings.HasPrefix(colName, "app_") {
				record, err := dbobj.store.GetRecordInTable(colName, "token", userTOKEN)
				if err != nil {
					return nil, err
				}
				if record != nil {
					results[colName[4:]] = record
				}
			}
		}
	} else {
		records, err := dbobj.store.GetList(storage.TblName.Userapps, "tone", userTOKEN, 0, 0, "appname")
		if err != nil {
			return nil, err
		}
		count := len(records)
		if count == 0 {
			return []byte("[]"), nil
		}
		for _, rec := range records {
			appname := rec["appname"].(string)
			delete(rec, "appname")
			results[appname] = rec
		}
	}
	if len(results) == 0 {
		return nil, nil
	}
	return json.Marshal(results)
}

func (dbobj dbcon) listAllAppsOnly(conf Config) ([]string, error) {
	var result []string
	if conf.Generic.UseSeparateAppTables == true {
		allCollections, err := dbobj.store.GetAllTables()
		if err != nil {
			return nil, err
		}
		for _, colName := range allCollections {
			if strings.HasPrefix(colName, "app_") {
				result = append(result, colName[4:])
			}
		}
	} else {
		records, err := dbobj.store.GetUniqueList(storage.TblName.Userapps, "appname")
		if err != nil {
			return result, err
		}
		for _, rec := range records {
			appname := rec["appname"].(string)
			result = append(result, appname)
		}
	}
	return result, nil
}

func (dbobj dbcon) listAllApps(conf Config) ([]byte, error) {
	result, err := dbobj.listAllAppsOnly(conf)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
