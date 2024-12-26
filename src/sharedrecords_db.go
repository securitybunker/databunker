package main

import (
	"errors"
	"log"
	"strings"
	"time"

	uuid "github.com/hashicorp/go-uuid"
	"github.com/securitybunker/databunker/src/storage"
	"github.com/securitybunker/databunker/src/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbobj dbcon) saveSharedRecord(userTOKEN string, fields string, expiration string, session string, appName string, partner string, conf Config) (string, error) {
	if utils.CheckValidUUID(userTOKEN) == false {
		return "", errors.New("bad uuid")
	}
	if len(expiration) == 0 {
		return "", errors.New("failed to parse expiration")
	}
	if len(appName) > 0 {
		apps, _ := dbobj.listAllApps(conf)
		if strings.Contains(string(apps), appName) == false {
			return "", errors.New("app not found")
		}
	}

	log.Printf("Expiration is : %s\n", expiration)
	start, err := utils.ParseExpiration(expiration)
	if err != nil {
		return "", err
	}
	// check if user record exists
	record, err := dbobj.lookupUserRecord(userTOKEN)
	if record == nil || err != nil {
		// not found
		return "", errors.New("not found")
	}
	recordUUID, err := uuid.GenerateUUID()
	if err != nil {
		return "", err
	}
	now := int32(time.Now().Unix())
	bdoc := bson.M{}
	bdoc["token"] = userTOKEN
	bdoc["record"] = recordUUID
	bdoc["when"] = now
	bdoc["endtime"] = start
	if len(fields) > 0 {
		bdoc["fields"] = fields
	}
	if len(appName) > 0 {
		bdoc["app"] = appName
	}
	if len(partner) > 0 {
		bdoc["partner"] = partner
	}
	if len(session) > 0 {
		bdoc["session"] = session
	}
	_, err = dbobj.store.CreateRecord(storage.TblName.Sharedrecords, &bdoc)
	if err != nil {
		return "", err
	}
	return recordUUID, nil
}

func (dbobj dbcon) getSharedRecord(recordUUID string) (checkRecordResult, error) {
	var result checkRecordResult
	//if utils.CheckValidUUID(recordUUID) == false {
	//	return result, errors.New("failed to authenticate")
	//}
	record, err := dbobj.store.GetRecord(storage.TblName.Sharedrecords, "record", recordUUID)
	if record == nil || err != nil {
		return result, errors.New("failed to authenticate")
	}
	result.name = recordUUID
	// tokenType = temp
	now := int32(time.Now().Unix())
	if now > record["endtime"].(int32) {
		return result, errors.New("xtoken expired")
	}
	result.token = record["token"].(string)
	if value, ok := record["fields"]; ok {
		result.fields = value.(string)
	}
	if value, ok := record["session"]; ok {
		result.session = value.(string)
	}
	if value, ok := record["app"]; ok {
		result.appName = value.(string)
	}

	return result, nil
}
