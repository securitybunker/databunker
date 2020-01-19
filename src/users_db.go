package databunker

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	uuid "github.com/hashicorp/go-uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbobj dbcon) createUserRecord(parsedData userJSON, event *auditEvent) (string, error) {
	var userTOKEN string
	//var bdoc interface{}
	bdoc := bson.M{}
	userTOKEN, err := uuid.GenerateUUID()
	if err != nil {
		return "", err
	}
	recordKey, err := generateRecordKey()
	if err != nil {
		return "", err
	}
	//err = bson.UnmarshalExtJSON(jsonData, false, &bdoc)
	encoded, err := encrypt(dbobj.masterKey, recordKey, parsedData.jsonData)
	if err != nil {
		return "", err
	}
	encodedStr := base64.StdEncoding.EncodeToString(encoded)
	fmt.Printf("data %s %s\n", parsedData.jsonData, encodedStr)
	bdoc["key"] = base64.StdEncoding.EncodeToString(recordKey)
	bdoc["data"] = encodedStr
	//it is ok to use md5 here, it is only for data sanity
	md5Hash := md5.Sum([]byte(encodedStr))
	bdoc["md5"] = base64.StdEncoding.EncodeToString(md5Hash[:])
	bdoc["token"] = userTOKEN
	// the index search field is hashed here, to be not-reversible
	// I use original md5(master_key) as a kind of salt here,
	// so no additional configuration field is needed here.
	if len(parsedData.loginIdx) > 0 {
		bdoc["loginidx"] = hashString(dbobj.hash, parsedData.loginIdx)
	}
	if len(parsedData.emailIdx) > 0 {
		bdoc["emailidx"] = hashString(dbobj.hash, parsedData.emailIdx)
	}
	if len(parsedData.phoneIdx) > 0 {
		bdoc["phoneidx"] = hashString(dbobj.hash, parsedData.phoneIdx)
	}
	if event != nil {
		event.After = encodedStr
		event.Record = userTOKEN
	}
	//fmt.Println("creating new user")
	_, err = dbobj.createRecord(TblName.Users, bdoc)
	if err != nil {
		fmt.Printf("error in create!\n")
		return "", err
	}
	return userTOKEN, nil
}

func (dbobj dbcon) generateTempLoginCode(userTOKEN string) int32 {
	rnd := randNum(4)
	fmt.Printf("random: %d\n", rnd)
	bdoc := bson.M{}
	bdoc["tempcode"] = rnd
	expired := int32(time.Now().Unix()) + 60
	bdoc["tempcodeexp"] = expired
	//fmt.Printf("op json: %s\n", update)
	dbobj.updateRecord(TblName.Users, "token", userTOKEN, &bdoc)
	return rnd
}

func (dbobj dbcon) generateDemoLoginCode(userTOKEN string) int32 {
	rnd := int32(4444)
	fmt.Printf("random: %d\n", rnd)
	bdoc := bson.M{}
	bdoc["tempcode"] = 4444
	expired := int32(time.Now().Unix()) + 60
	bdoc["tempcodeexp"] = expired
	//fmt.Printf("op json: %s\n", update)
	dbobj.updateRecord(TblName.Users, "token", userTOKEN, &bdoc)
	return rnd
}

// int 0 - same value
// int -1 remove
// int 1 add
func (dbobj dbcon) validateIndexChange(indexName string, idxOldValue string, raw map[string]interface{}, conf Config) (int, error) {
	if len(idxOldValue) == 0 {
		return 0, nil
	}
	// check type of raw[indexName]
	//fmt.Println(raw[indexName])
	if newIdxValue, ok2 := raw[indexName]; ok2 {
		if reflect.TypeOf(newIdxValue) == reflect.TypeOf("string") {
			idxStringHashHex := hashString(dbobj.hash, newIdxValue.(string))
			if idxStringHashHex != idxOldValue {
				// old index value renamed
				// check if this value is uniqueue
				otherUserBson, _ := dbobj.lookupUserRecordByIndex(indexName, newIdxValue.(string), conf)
				if otherUserBson != nil {
					// already exist user with same index value
					return 0, errors.New("duplicate index")
				}
				//fmt.Println("new index value good")
				return 1, nil
			}
			// same value, no need to check
			//fmt.Println("same index value")
			return 0, nil
		} else if reflect.TypeOf(newIdxValue) == reflect.TypeOf(nil) {
			//fmt.Println("old index removed!!!")
			return -1, nil
		} else {
			// index value is changed to unknown value type
			//e := fmt.Sprintf("wrong index type for %s : %s", indexName, reflect.TypeOf(newIdxValue))
			//return 0, errors.New(e)
			// silently remove index as value is not string
			return -1, nil
		}
	}
	// index value removed
	//fmt.Println("old index removed!")
	return -1, nil
}

func (dbobj dbcon) updateUserRecord(parsedData userJSON, userTOKEN string, event *auditEvent, conf Config) ([]byte, []byte, error) {
	var err error
	for x := 0; x < 10; x++ {
		oldJSON, newJSON, err := dbobj.updateUserRecordDo(parsedData, userTOKEN, event, conf)
		if err == nil {
			return oldJSON, newJSON, nil
		}
		fmt.Printf("Trying to update user again: %s\n", userTOKEN)
	}
	return nil, nil, err
}

func (dbobj dbcon) updateUserRecordDo(parsedData userJSON, userTOKEN string, event *auditEvent, conf Config) ([]byte, []byte, error) {
	//_, err = collection.InsertOne(context.TODO(), bson.M{"name": "The Go Language2", "genre": "Coding", "authorId": "4"})
	oldUserBson, err := dbobj.lookupUserRecord(userTOKEN)
	if oldUserBson == nil || err != nil {
		// not found
		return nil, nil, err
	}

	// get user key
	userKey := oldUserBson["key"].(string)
	recordKey, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return nil, nil, err
	}
	encData0 := oldUserBson["data"].(string)
	encData, err := base64.StdEncoding.DecodeString(encData0)
	if err != nil {
		return nil, nil, err
	}
	decrypted, err := decrypt(dbobj.masterKey, recordKey, encData)
	if err != nil {
		return nil, nil, err
	}
	// merge
	fmt.Printf("old json: %s\n", decrypted)
	jsonDataPatch := parsedData.jsonData
	fmt.Printf("json patch: %s\n", jsonDataPatch)
	newJSON, err := jsonpatch.MergePatch(decrypted, jsonDataPatch)
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("result: %s\n", newJSON)

	var raw map[string]interface{}
	err = json.Unmarshal(newJSON, &raw)
	if err != nil {
		return nil, nil, err
	}
	bdel := bson.M{}
	sig := oldUserBson["md5"].(string)
	// create new user record
	bdoc := bson.M{}
	keys := []string{"login", "email", "phone"}
	for _, idx := range keys {
		//fmt.Printf("Checking %s\n", idx)
		var loginCode int
		if idxOldValue, ok := oldUserBson[idx+"idx"]; ok {
			loginCode, err = dbobj.validateIndexChange(idx, idxOldValue.(string), raw, conf)
			if err != nil {
				return nil, nil, err
			}
			if loginCode == -1 {
				bdel[idx+"idx"] = ""
			}
		} else {
			// check if new value is created
			if newIdxValue, ok3 := raw[idx]; ok3 {
				//fmt.Printf("adding index? %s\n", raw[idx])
				otherUserBson, _ := dbobj.lookupUserRecordByIndex(idx, newIdxValue.(string), conf)
				if otherUserBson != nil {
					// already exist user with same index value
					return nil, nil, fmt.Errorf("duplicate %s index", idx)
				}
				//fmt.Printf("adding index2? %s\n", raw[idx])
				// create login index
				loginCode = 1
			}
		}
		if loginCode == 1 {
			//fmt.Printf("adding index3? %s\n", raw[idx])
			bdoc[idx+"idx"] = hashString(dbobj.hash, raw[idx].(string))
		}
	}

	encoded, _ := encrypt(dbobj.masterKey, recordKey, newJSON)
	encodedStr := base64.StdEncoding.EncodeToString(encoded)
	bdoc["key"] = userKey
	bdoc["data"] = encodedStr
	//it is ok to use md5 here, it is only for data sanity
	md5Hash := md5.Sum([]byte(encodedStr))
	bdoc["md5"] = base64.StdEncoding.EncodeToString(md5Hash[:])
	bdoc["token"] = userTOKEN

	// here I add md5 of the original record to filter
	// to make sure this record was not change by other thread
	//filter2 := bson.D{{"token", userTOKEN}, {"md5", sig}}

	//fmt.Printf("op json: %s\n", update)
	result, err := dbobj.updateRecord2(TblName.Users, "token", userTOKEN, "md5", sig, &bdoc, &bdel)
	if err != nil {
		return nil, nil, err
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
	return decrypted, newJSON, nil
}

func (dbobj dbcon) lookupUserRecord(userTOKEN string) (bson.M, error) {
	return dbobj.getRecord(TblName.Users, "token", userTOKEN)
}

func (dbobj dbcon) lookupUserRecordByIndex(indexName string, indexValue string, conf Config) (bson.M, error) {
	if indexName == "email" {
		indexValue = normalizeEmail(indexValue)
	} else if indexName == "phone" {
		indexValue = normalizePhone(indexValue, conf.Sms.DefaultCountry)
	}
	if len(indexValue) == 0 {
		return nil, nil
	}
	idxStringHashHex := hashString(dbobj.hash, indexValue)
	fmt.Printf("loading by %s\n", indexName)
	return dbobj.getRecord(TblName.Users, indexName+"idx", idxStringHashHex)
}

func (dbobj dbcon) getUser(userTOKEN string) ([]byte, error) {
	userBson, err := dbobj.lookupUserRecord(userTOKEN)
	if userBson == nil || err != nil {
		// not found
		return nil, err
	}
	userKey := userBson["key"].(string)
	recordKey, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return nil, err
	}
	var decrypted []byte
	if _, ok := userBson["data"]; ok {
		encData0 := userBson["data"].(string)
		if len(encData0) > 0 {
			encData, err := base64.StdEncoding.DecodeString(encData0)
			if err != nil {
				return nil, err
			}
			decrypted, err = decrypt(dbobj.masterKey, recordKey, encData)
			if err != nil {
				return nil, err
			}
		}
	}
	return decrypted, err
}

func (dbobj dbcon) getUserIndex(indexValue string, indexName string, conf Config) ([]byte, string, error) {
	userBson, err := dbobj.lookupUserRecordByIndex(indexName, indexValue, conf)
	if userBson == nil || err != nil {
		return nil, "", err
	}
	// decrypt record
	userKey := userBson["key"].(string)
	recordKey, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return nil, "", err
	}
	var decrypted []byte
	if _, ok := userBson["data"]; ok {
		encData0 := userBson["data"].(string)
		if len(encData0) > 0 {
			encData, err := base64.StdEncoding.DecodeString(encData0)
			if err != nil {
				return nil, "", err
			}
			decrypted, err = decrypt(dbobj.masterKey, recordKey, encData)
			if err != nil {
				return nil, "", err
			}
		}
	}
	return decrypted, userBson["token"].(string), err
}

func (dbobj dbcon) deleteUserRecord(userTOKEN string) (bool, error) {
	userApps, err := dbobj.listAllAppsOnly()
	if err != nil {
		return false, err
	}
	// delete all user app records
	for _, appName := range userApps {
		appNameFull := "app_" + appName
		dbobj.deleteRecordInTable(appNameFull, "token", userTOKEN)
	}
	//delete in audit
	dbobj.deleteRecordInTable("audit", "record", userTOKEN)
	dbobj.deleteRecordInTable("sessions", "token", userTOKEN)
	// cleanup user record
	bdel := bson.M{}
	bdel["data"] = ""
	bdel["key"] = ""
	bdel["loginidx"] = ""
	bdel["emailidx"] = ""
	bdel["phoneidx"] = ""
	result, err := dbobj.cleanupRecord(TblName.Users, "token", userTOKEN, bdel)
	if err != nil {
		return false, err
	}
	if result > 0 {
		return true, nil
	}
	return true, nil
}

func (dbobj dbcon) wipeRecord(userTOKEN string) (bool, error) {
	userApps, err := dbobj.listAllAppsOnly()
	if err != nil {
		return false, err
	}
	// delete all user app records
	for _, appName := range userApps {
		appNameFull := "app_" + appName
		dbobj.deleteRecordInTable(appNameFull, "token", userTOKEN)
	}
	// delete user record
	result, err := dbobj.deleteRecord(TblName.Users, "token", userTOKEN)
	if err != nil {
		return false, err
	}
	if result > 0 {
		return true, nil
	}
	return false, nil
}

func (dbobj dbcon) userEncrypt(userTOKEN string, data []byte) (string, error) {
	userBson, err := dbobj.lookupUserRecord(userTOKEN)
	if err != nil {
		// not found
		return "", errors.New("not found")
	}
	if userBson == nil {
		return "", errors.New("not found")
	}
	userKey := userBson["key"].(string)
	recordKey, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return "", err
	}
	// encrypt data
	encoded, err := encrypt(dbobj.masterKey, recordKey, data)
	if err != nil {
		return "", err
	}
	encodedStr := base64.StdEncoding.EncodeToString(encoded)
	return encodedStr, nil
}

func (dbobj dbcon) userDecrypt(userTOKEN, src string) ([]byte, error) {
	userBson, err := dbobj.lookupUserRecord(userTOKEN)
	if err != nil {
		// not found
		return nil, errors.New("not found")
	}
	if userBson == nil {
		return nil, errors.New("not found")
	}
	userKey := userBson["key"].(string)
	recordKey, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return nil, err
	}
	encData, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return nil, err
	}
	decrypted, err := decrypt(dbobj.masterKey, recordKey, encData)
	return decrypted, err
}

func (dbobj dbcon) userDecrypt2(userTOKEN, src string, src2 string) ([]byte, []byte, error) {
	userBson, err := dbobj.lookupUserRecord(userTOKEN)
	if err != nil {
		// not found
		return nil, nil, errors.New("not found")
	}
	if userBson == nil {
		return nil, nil, errors.New("not found")
	}
	userKey := userBson["key"].(string)
	recordKey, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return nil, nil, err
	}
	encData, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return nil, nil, err
	}
	decrypted, err := decrypt(dbobj.masterKey, recordKey, encData)
	if len(src2) == 0 {
		return decrypted, nil, err
	}
	encData2, err := base64.StdEncoding.DecodeString(src2)
	if err != nil {
		return decrypted, nil, err
	}
	decrypted2, err := decrypt(dbobj.masterKey, recordKey, encData2)
	return decrypted, decrypted2, err
}
