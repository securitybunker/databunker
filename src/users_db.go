package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/securitybunker/databunker/src/storage"
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
	userKeyBinary, err := generateRecordKey()
	if err != nil {
		return "", err
	}
	//err = bson.UnmarshalExtJSON(jsonData, false, &bdoc)
	encoded, err := encrypt(dbobj.masterKey, userKeyBinary, parsedData.jsonData)
	if err != nil {
		return "", err
	}
	encodedStr := base64.StdEncoding.EncodeToString(encoded)
	//fmt.Printf("data %s %s\n", parsedData.jsonData, encodedStr)
	bdoc["key"] = base64.StdEncoding.EncodeToString(userKeyBinary)
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
	if len(parsedData.customIdx) > 0 {
		bdoc["customidx"] = hashString(dbobj.hash, parsedData.customIdx)
	}
	if event != nil {
		event.After = encodedStr
		event.Record = userTOKEN
	}
	//fmt.Println("creating new user")
	_, err = dbobj.store.CreateRecord(storage.TblName.Users, &bdoc)
	if err != nil {
		fmt.Printf("error in create!\n")
		return "", err
	}
	return userTOKEN, nil
}

func (dbobj dbcon) initiateUserExpiration(userTOKEN string, endtime int32, status string, expToken string) error {
	bdoc := bson.M{}
	bdoc["endtime"] = endtime
	bdoc["expstatus"] = status
	bdoc["exptoken"] = expToken
	_, err := dbobj.store.UpdateRecord(storage.TblName.Users, "token", userTOKEN, &bdoc)
	return err
}

func (dbobj dbcon) updateUserExpStatus(userTOKEN string, status string) error {
	bdoc := bson.M{}
	if status == "retain" {
		bdoc["endtime"] = 0
		bdoc["exptoken"] = ""
	} else if len(status) == 0 {
		// cancel expiration
		bdoc["endtime"] = 0
		bdoc["exptoken"] = ""
	}
	bdoc["expstatus"] = status
	_, err := dbobj.store.UpdateRecord(storage.TblName.Users, "token", userTOKEN, &bdoc)
	return err
}

func (dbobj dbcon) generateTempLoginCode(userTOKEN string) int32 {
	rnd := randNum(6)
	fmt.Printf("random: %d\n", rnd)
	bdoc := bson.M{}
	bdoc["tempcode"] = rnd
	expired := int32(time.Now().Unix()) + 60
	bdoc["tempcodeexp"] = expired
	//fmt.Printf("op json: %s\n", update)
	dbobj.store.UpdateRecord(storage.TblName.Users, "token", userTOKEN, &bdoc)
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
	dbobj.store.UpdateRecord(storage.TblName.Users, "token", userTOKEN, &bdoc)
	return rnd
}

func (dbobj dbcon) validateUserRecordChange(oldUserJSON []byte, jsonDataPatch []byte, userTOKEN string, authResult string) (bool, error) {
	// prepare merge
	//fmt.Printf("old json: %s\n", oldUserJSON)
	//fmt.Printf("json patch: %s\n", jsonDataPatch)
	var newJSON []byte
	var err error
	if jsonDataPatch[0] == '{' {
		newJSON, err = jsonpatch.MergePatch(oldUserJSON, jsonDataPatch)
	} else {
		patch, err := jsonpatch.DecodePatch(jsonDataPatch)
		if err != nil {
			return false, err
		}
		newJSON, err = patch.Apply(oldUserJSON)
	}
	if err != nil {
		return false, err
	}
	//fmt.Printf("result: %s\n", newJSON)
	return validateUserRecordChange(oldUserJSON, newJSON, authResult)
}

func (dbobj dbcon) updateUserRecord(jsonDataPatch []byte, userTOKEN string, userBSON bson.M, event *auditEvent, conf Config) ([]byte, []byte, bool, error) {
	oldJSON, newJSON, lookupErr, err := dbobj.updateUserRecordDo(jsonDataPatch, userTOKEN, userBSON, event, conf)
	if lookupErr == true {
		return oldJSON, newJSON, lookupErr, err
	}
	if err == nil {
		return oldJSON, newJSON, lookupErr, nil
	}
	// load one more time user BSON structure
	userBSON2, err := dbobj.lookupUserRecord(userTOKEN)
	if userBSON2 == nil || err != nil {
		return nil, nil, true, err
	}
	oldJSON, newJSON, lookupErr, err = dbobj.updateUserRecordDo(jsonDataPatch, userTOKEN, userBSON2, event, conf)
	if lookupErr == true {
		return oldJSON, newJSON, lookupErr, err
	}
	if err == nil {
		return oldJSON, newJSON, lookupErr, nil
	}
	return nil, nil, false, err
}

func (dbobj dbcon) updateUserRecordDo(jsonDataPatch []byte, userTOKEN string, oldUserBson bson.M, event *auditEvent, conf Config) ([]byte, []byte, bool, error) {

	// get user key
	userKey := oldUserBson["key"].(string)
	userKeyBinary, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return nil, nil, false, err
	}
	encData0 := oldUserBson["data"].(string)
	encData, err := base64.StdEncoding.DecodeString(encData0)
	if err != nil {
		return nil, nil, false, err
	}
	decrypted, err := decrypt(dbobj.masterKey, userKeyBinary, encData)
	if err != nil {
		return nil, nil, false, err
	}
	var raw2 map[string]interface{}
	err = json.Unmarshal(decrypted, &raw2)
	if err != nil {
		return nil, nil, false, err
	}
	oldEmail := ""
	if _, ok := raw2["email"]; ok {
		oldEmail = normalizeEmail(raw2["email"].(string))
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
			return nil, nil, false, err
		}
		newJSON, err = patch.Apply(decrypted)
	}
	if err != nil {
		return nil, nil, false, err
	}
	//fmt.Printf("result: %s\n", newJSON)

	var raw map[string]interface{}
	err = json.Unmarshal(newJSON, &raw)
	if err != nil {
		return nil, nil, false, err
	}
	bdel := bson.M{}
	sig := oldUserBson["md5"].(string)
	// create new user record
	bdoc := bson.M{}
	keys := []string{"login", "email", "phone", "custom"}
	newEmail := ""
	for _, idx := range keys {
		//fmt.Printf("Checking %s\n", idx)
		actionCode := 1
		newIdxFinalValue := ""
		if newIdxValue, ok3 := raw[idx]; ok3 {
			newIdxFinalValue = getIndexString(newIdxValue)
			//fmt.Println("newIdxFinalValue0", newIdxFinalValue)
			if len(newIdxFinalValue) > 0 {
				if idx == "email" {
					newIdxFinalValue = normalizeEmail(newIdxFinalValue)
					newEmail = newIdxFinalValue
				} else if idx == "phone" {
					newIdxFinalValue = normalizePhone(newIdxFinalValue, conf.Sms.DefaultCountry)
				}
			}
			//fmt.Println("newIdxFinalValue", newIdxFinalValue)
		}
		if idxOldValue, ok := oldUserBson[idx+"idx"]; ok {
			if len(newIdxFinalValue) > 0 && len(idxOldValue.(string)) >= 0 {
				idxStringHashHex := hashString(dbobj.hash, newIdxFinalValue)
				if idxStringHashHex == idxOldValue.(string) {
					//fmt.Println("index value NOT changed!")
					actionCode = 0
				} else {
					//fmt.Println("index value changed!")
				}
				//} else {
				//	fmt.Println("old or new is empty")
			}
		}
		if len(newIdxFinalValue) > 0 && actionCode == 1 {
			// check if new value is created
			//fmt.Printf("adding index? %s\n", raw[idx])
			otherUserBson, _ := dbobj.lookupUserRecordByIndex(idx, newIdxFinalValue, conf)
			if otherUserBson != nil {
				// already exist user with same index value
				return nil, nil, true, fmt.Errorf("duplicate %s index", idx)
			}
			//fmt.Printf("adding index3? %s\n", raw[idx])
			bdoc[idx+"idx"] = hashString(dbobj.hash, newIdxFinalValue)
		} else if len(newIdxFinalValue) == 0 {
			bdel[idx+"idx"] = ""
		}
	}

	encoded, _ := encrypt(dbobj.masterKey, userKeyBinary, newJSON)
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
	result, err := dbobj.store.UpdateRecord2(storage.TblName.Users, "token", userTOKEN, "md5", sig, &bdoc, &bdel)
	if err != nil {
		return nil, nil, false, err
	}
	if oldEmail != newEmail && len(oldEmail) > 0 && len(newEmail) > 0 {
		dbobj.GlobalUserChangeEmail(oldEmail, newEmail)
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
	return decrypted, newJSON, false, nil
}

func (dbobj dbcon) lookupUserRecord(userTOKEN string) (bson.M, error) {
	return dbobj.store.GetRecord(storage.TblName.Users, "token", userTOKEN)
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
	if indexName == "exptoken" {
		return dbobj.store.GetRecord(storage.TblName.Users, "exptoken", indexValue)
	}
	idxStringHashHex := hashString(dbobj.hash, indexValue)
	//fmt.Printf("loading by %s, value: %s\n", indexName, indexValue)
	return dbobj.store.GetRecord(storage.TblName.Users, indexName+"idx", idxStringHashHex)
}

func (dbobj dbcon) getUserJSON(userTOKEN string) ([]byte, error) {
	userBson, err := dbobj.lookupUserRecord(userTOKEN)
	if userBson == nil || err != nil {
		// not found
		return nil, err
	}
	if _, ok := userBson["key"]; !ok {
		return []byte("{}"), nil
	}
	userKey := userBson["key"].(string)
	userKeyBinary, err := base64.StdEncoding.DecodeString(userKey)
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
			decrypted, err = decrypt(dbobj.masterKey, userKeyBinary, encData)
			if err != nil {
				return nil, err
			}
		}
	}
	return decrypted, err
}

func (dbobj dbcon) getUser(userTOKEN string) ([]byte, bson.M, error) {
	userBson, err := dbobj.lookupUserRecord(userTOKEN)
	if userBson == nil || err != nil {
		// not found
		return nil, nil, err
	}
	if _, ok := userBson["key"]; !ok {
		return []byte("{}"), userBson, nil
	}
	userKey := userBson["key"].(string)
	userKeyBinary, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return nil, nil, err
	}
	var decrypted []byte
	if _, ok := userBson["data"]; ok {
		encData0 := userBson["data"].(string)
		if len(encData0) > 0 {
			encData, err := base64.StdEncoding.DecodeString(encData0)
			if err != nil {
				return nil, nil, err
			}
			decrypted, err = decrypt(dbobj.masterKey, userKeyBinary, encData)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	return decrypted, userBson, err
}


func (dbobj dbcon) dumpUserPII(email string, conf Config) (string, error) {
	fullJSON := ""
	profileJSON, userTOKEN, err := dbobj.getUserJSONByIndex(email, "email", conf)
	if userTOKEN != "" {
		fullJSON = fmt.Sprintf(`{"profile":%s`, profileJSON)
		userappsJSON, _ := dbobj.dumpUserApps(userTOKEN, conf)
		if userappsJSON != nil {
			fullJSON += fmt.Sprintf(`,"apps":%s`, userappsJSON)
		}
		agreementsJSON, numAgreements, _ := dbobj.listAgreementRecords(userTOKEN)
		if numAgreements > 0 {
			fullJSON += fmt.Sprintf(`,"agreements":%s`, agreementsJSON)
		}
		fullJSON += "}"
	} else {
		agreementsJSON, numAgreements, _ := dbobj.listAgreementRecordsByIdentity(email)
		if numAgreements > 0 {
			fullJSON = fmt.Sprintf(`{"agreements":%s}`, agreementsJSON)
		}
	}
	return fullJSON, err
}


func (dbobj dbcon) getUserJSONByIndex(indexValue string, indexName string, conf Config) ([]byte, string, error) {
	userBson, err := dbobj.lookupUserRecordByIndex(indexName, indexValue, conf)
	if userBson == nil || err != nil {
		return nil, "", err
	}
	// decrypt record
	userKey := userBson["key"].(string)
	userKeyBinary, err := base64.StdEncoding.DecodeString(userKey)
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
			decrypted, err = decrypt(dbobj.masterKey, userKeyBinary, encData)
			if err != nil {
				return nil, "", err
			}
		}
	}
	return decrypted, userBson["token"].(string), err
}

func (dbobj dbcon) getUserByIndex(indexValue string, indexName string, conf Config) ([]byte, string, bson.M, error) {
	userBson, err := dbobj.lookupUserRecordByIndex(indexName, indexValue, conf)
	if userBson == nil || err != nil {
		return nil, "", nil, err
	}
	// decrypt record
	userKey := userBson["key"].(string)
	userKeyBinary, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return nil, "", nil, err
	}
	var decrypted []byte
	if _, ok := userBson["data"]; ok {
		encData0 := userBson["data"].(string)
		if len(encData0) > 0 {
			encData, err := base64.StdEncoding.DecodeString(encData0)
			if err != nil {
				return nil, "", nil, err
			}
			decrypted, err = decrypt(dbobj.masterKey, userKeyBinary, encData)
			if err != nil {
				return nil, "", nil, err
			}
		}
	}
	return decrypted, userBson["token"].(string), userBson, err
}

func (dbobj dbcon) deleteUserRecord(userJSON []byte, userTOKEN string, conf Config) (bool, error) {
	dbobj.deleteUserApps(userTOKEN, conf)
	//delete in audit
	dbobj.store.DeleteRecord(storage.TblName.Audit, "record", userTOKEN)
	dbobj.store.DeleteRecord(storage.TblName.Sessions, "token", userTOKEN)

	dataJSON, record := cleanupRecord(userJSON)
	bdel := bson.M{}
	if dataJSON != nil {
		oldUserBson, err := dbobj.lookupUserRecord(userTOKEN)
		if err != nil {
			return false, err
		}
		userKey := oldUserBson["key"].(string)
		userKeyBinary, err := base64.StdEncoding.DecodeString(userKey)
		if err != nil {
			return false, err
		}
		sig := oldUserBson["md5"].(string)
		bdoc := bson.M{}

		if _, ok := record["email"]; ok {
			fmt.Printf("Preservice email idx\n")
			bdoc["emailidx"] = oldUserBson["emailidx"].(string)
		} else {
			bdel["emailidx"] = ""
		}
		if _, ok := record["phone"]; ok {
			fmt.Printf("Preservice phone idx\n")
			bdoc["phoneidx"] = oldUserBson["phoneidx"].(string)
		} else {
			bdel["phoneidx"] = ""
		}
		if _, ok := record["login"]; ok {
			fmt.Printf("Preservice login idx\n")
			bdoc["loginidx"] = oldUserBson["loginidx"].(string)
		} else {
			bdel["loginidx"] = ""
		}
		encoded, _ := encrypt(dbobj.masterKey, userKeyBinary, dataJSON)
		encodedStr := base64.StdEncoding.EncodeToString(encoded)
		bdoc["key"] = userKey
		bdoc["data"] = encodedStr
		md5Hash := md5.Sum([]byte(encodedStr))
		bdoc["md5"] = base64.StdEncoding.EncodeToString(md5Hash[:])
		bdoc["token"] = userTOKEN
		result, err := dbobj.store.UpdateRecord2(storage.TblName.Users, "token", userTOKEN, "md5", sig, &bdoc, &bdel)
		if err != nil {
			return false, err
		}
		if result > 0 {
			return true, nil
		}
		return false, nil
	}
	// cleanup user record
	bdel["data"] = ""
	bdel["key"] = ""
	bdel["loginidx"] = ""
	bdel["emailidx"] = ""
	bdel["phoneidx"] = ""
	bdel["customidx"] = ""
	result, err := dbobj.store.CleanupRecord(storage.TblName.Users, "token", userTOKEN, bdel)
	if err != nil {
		return false, err
	}
	if result > 0 {
		return true, nil
	}
	return true, nil
}

/*
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
	result, err := dbobj.store.DeleteRecord(storage.TblName.Users, "token", userTOKEN)
	if err != nil {
		return false, err
	}
	if result > 0 {
		return true, nil
	}
	return false, nil
}
*/

func (dbobj dbcon) userEncrypt(userTOKEN string, data []byte) (string, error) {
	userBson, err := dbobj.lookupUserRecord(userTOKEN)
	if userBson == nil || err != nil {
		return "", errors.New("not found")
	}
	if _, ok := userBson["key"]; !ok {
		// user might be deleted already
		return "", errors.New("not found")
	}
	userKey := userBson["key"].(string)
	userKeyBinary, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return "", err
	}
	// encrypt data
	encoded, err := encrypt(dbobj.masterKey, userKeyBinary, data)
	if err != nil {
		return "", err
	}
	encodedStr := base64.StdEncoding.EncodeToString(encoded)
	return encodedStr, nil
}

func (dbobj dbcon) userDecrypt(userTOKEN, src string) ([]byte, error) {
	userBson, err := dbobj.lookupUserRecord(userTOKEN)
	if userBson == nil || err != nil {
		return nil, errors.New("not found")
	}
	if _, ok := userBson["key"]; !ok {
		// user might be deleted already
		return nil, errors.New("not found")
	}
	userKey := userBson["key"].(string)
	userKeyBinary, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return nil, err
	}
	encData, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return nil, err
	}
	decrypted, err := decrypt(dbobj.masterKey, userKeyBinary, encData)
	return decrypted, err
}

func (dbobj dbcon) userDecrypt2(userTOKEN, src string, src2 string) ([]byte, []byte, error) {
	userBson, err := dbobj.lookupUserRecord(userTOKEN)
	if userBson == nil || err != nil {
		return nil, nil, errors.New("not found")
	}
	if _, ok := userBson["key"]; !ok {
		// user might be deleted already
		return nil, nil, errors.New("not found")
	}
	userKey := userBson["key"].(string)
	userKeyBinary, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return nil, nil, err
	}
	encData, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return nil, nil, err
	}
	decrypted, err := decrypt(dbobj.masterKey, userKeyBinary, encData)
	if len(src2) == 0 {
		return decrypted, nil, err
	}
	encData2, err := base64.StdEncoding.DecodeString(src2)
	if err != nil {
		return decrypted, nil, err
	}
	decrypted2, err := decrypt(dbobj.masterKey, userKeyBinary, encData2)
	return decrypted, decrypted2, err
}
