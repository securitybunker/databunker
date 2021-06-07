package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/securitybunker/databunker/src/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type sessionEvent struct {
	When int32  `json:"when"`
	Data string `json:"data"`
}

func (dbobj dbcon) createSessionRecord(sessionUUID string, userTOKEN string, expiration string, data []byte) (string, error) {
	var endtime int32
	var err error
	if len(expiration) > 0 {
		endtime, err = parseExpiration(expiration)
		if err != nil {
			return "", err
		}
	}
	recordKey, err := generateRecordKey()
	if err != nil {
		return "", err
	}
	encoded, err := encrypt(dbobj.masterKey, recordKey, data)
	if err != nil {
		return "", err
	}
	encodedStr := base64.StdEncoding.EncodeToString(encoded)
	bdoc := bson.M{}
	now := int32(time.Now().Unix())
	bdoc["token"] = userTOKEN
	bdoc["session"] = sessionUUID
	bdoc["endtime"] = endtime
	bdoc["when"] = now
	bdoc["data"] = encodedStr
	bdoc["key"] = base64.StdEncoding.EncodeToString(recordKey)
	record, err := dbobj.store.GetRecord(storage.TblName.Sessions, "session", sessionUUID)
	if record == nil || len(record) == 0 {
		_, err = dbobj.store.CreateRecord(storage.TblName.Sessions, &bdoc)
		if err != nil {
			return "", err
		}
		return sessionUUID, nil
	}
	dbobj.store.UpdateRecord(storage.TblName.Sessions, "session", sessionUUID, &bdoc)
	return sessionUUID, nil
}

func (dbobj dbcon) getSession(sessionUUID string) (int32, []byte, string, error) {
	record, err := dbobj.store.GetRecord(storage.TblName.Sessions, "session", sessionUUID)
	if err != nil {
		return 0, nil, "", err
	}
	if record == nil {
		return 0, nil, "", errors.New("not found")
	}
	// check expiration
	now := int32(time.Now().Unix())
	if now > record["endtime"].(int32) {
		return 0, nil, "", errors.New("session expired")
	}
	when := record["when"].(int32)
	userTOKEN := record["token"].(string)
	encData0 := record["data"].(string)
	recordKey0 := record["key"].(string)
	recordKey, err := base64.StdEncoding.DecodeString(recordKey0)
	if err != nil {
		return 0, nil, "", err
	}
	encData, err := base64.StdEncoding.DecodeString(encData0)
	if err != nil {
		return 0, nil, "", err
	}
	decrypted, err := decrypt(dbobj.masterKey, recordKey, encData)
	if err != nil {
		return 0, nil, "", err
	}
	return when, decrypted, userTOKEN, err
}

func (dbobj dbcon) getUserSessionsByToken(userTOKEN string, offset int32, limit int32) ([]string, int64, error) {
	count, err := dbobj.store.CountRecords(storage.TblName.Sessions, "token", userTOKEN)
	if err != nil {
		return nil, 0, err
	}
	records, err := dbobj.store.GetList(storage.TblName.Sessions, "token", userTOKEN, offset, limit, "")
	if err != nil {
		return nil, 0, err
	}
	var results []string
	for _, element := range records {
		when := element["when"].(int32)
		session := element["session"].(string)
		encData0 := element["data"].(string)
		recordKey0 := element["key"].(string)
		recordKey, _ := base64.StdEncoding.DecodeString(recordKey0)
		encData, _ := base64.StdEncoding.DecodeString(encData0)
		decrypted, _ := decrypt(dbobj.masterKey, recordKey, encData)
		sEvent := fmt.Sprintf(`{"when":%d,"session":"%s","data":%s}`, when, session, string(decrypted))
		results = append(results, sEvent)
	}

	return results, count, err
}

func (dbobj dbcon) deleteSession(sessionUUID string) (bool, error) {
	dbobj.store.DeleteRecord(storage.TblName.Sessions, "session", sessionUUID)
	return true, nil
}
