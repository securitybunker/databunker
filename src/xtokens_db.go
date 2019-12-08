package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	uuid "github.com/hashicorp/go-uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbobj dbcon) getRootToken() (string, error) {
	record, err := dbobj.getRecord(TblName.Xtokens, "type", "root")
	if err != nil {
		return "", err
	}
	if record == nil {
		return "", nil
	}
	return record["xtoken"].(string), nil
}

func (dbobj dbcon) createRootToken() (string, error) {
	rootToken, err := dbobj.getRootToken()
	if len(rootToken) > 0 {
		return rootToken, nil
	}
	rootToken, err = uuid.GenerateUUID()
	if err != nil {
		return "", err
	}
	bdoc := bson.M{}
	bdoc["xtoken"] = rootToken
	bdoc["type"] = "root"
	_, err = dbobj.createRecord(TblName.Xtokens, bdoc)
	if err != nil {
		return rootToken, err
	}
	return rootToken, nil
}

func (dbobj dbcon) generateUserTempXToken(userTOKEN string, fields string, expiration string, appName string) (string, error) {
	if isValidUUID(userTOKEN) == false {
		return "", errors.New("bad uuid")
	}
	if len(expiration) == 0 {
		return "", errors.New("failed to parse expiration")
	}
	if len(appName) > 0 {
		apps, _ := dbobj.listAllApps()
		if strings.Contains(string(apps), appName) == false {
			return "", errors.New("app not found")
		}
	}

	start, err := parseExpiration(expiration)
	if err != nil {
		return "", err
	}

	// check if user record exists
	record, err := dbobj.lookupUserRecord(userTOKEN)
	if record == nil || err != nil {
		// not found
		return "", errors.New("not found")
	}

	tokenUUID, err := uuid.GenerateUUID()
	if err != nil {
		return "", err
	}
	bdoc := bson.M{}
	bdoc["token"] = userTOKEN
	bdoc["xtoken"] = tokenUUID
	bdoc["type"] = "temp"
	bdoc["fields"] = fields
	bdoc["endtime"] = start
	if len(appName) > 0 {
		bdoc["app"] = appName
	}
	_, err = dbobj.createRecord(TblName.Xtokens, bdoc)
	if err != nil {
		return "", err
	}
	return tokenUUID, nil
}

func (dbobj dbcon) generateUserLoginXToken(userTOKEN string) (string, error) {
	if isValidUUID(userTOKEN) == false {
		return "", errors.New("bad token format")
	}

	// check if user record exists
	record, err := dbobj.lookupUserRecord(userTOKEN)
	if record == nil || err != nil {
		// not found
		return "", errors.New("not found")
	}

	tokenUUID, err := uuid.GenerateUUID()
	if err != nil {
		return "", err
	}
	// by default login token for 30 minutes only
	expired := int32(time.Now().Unix()) + 10*60
	bdoc := bson.M{}
	bdoc["token"] = userTOKEN
	bdoc["xtoken"] = tokenUUID
	bdoc["type"] = "login"
	bdoc["endtime"] = expired
	_, err = dbobj.createRecord(TblName.Xtokens, bdoc)
	if err != nil {
		return "", err
	}
	return tokenUUID, nil
}

func (dbobj dbcon) checkToken(tokenUUID string) bool {
	//fmt.Printf("Token0 %s\n", tokenUUID)
	if isValidUUID(tokenUUID) == false {
		return false
	}
	record, err := dbobj.getRecord(TblName.Xtokens, "xtoken", tokenUUID)
	if record == nil || err != nil {
		return false
	}
	tokenType := record["type"].(string)
	if tokenType == "root" {
		return true
	}
	return false
}

func (dbobj dbcon) checkUserAuthXToken(xtokenUUID string) (tokenAuthResult, error) {
	var result tokenAuthResult
	if isValidUUID(xtokenUUID) == false {
		return result, errors.New("failed to authenticate")
	}
	record, err := dbobj.getRecord(TblName.Xtokens, "xtoken", xtokenUUID)
	if record == nil || err != nil {
		return result, errors.New("failed to authenticate")
	}
	tokenType := record["type"].(string)
	fmt.Printf("token type: %s\n", tokenType)
	if tokenType == "root" {
		// we have this admin user
		result.ttype = "root"
		result.name = "root"
		return result, nil
	}
	result.name = xtokenUUID
	// tokenType = temp
	now := int32(time.Now().Unix())
	if now > record["endtime"].(int32) {
		return result, errors.New("token expired")
	}
	result.token = record["token"].(string)
	if value, ok := record["fields"]; ok {
		result.fields = value.(string)
	}
	if tokenType == "login" {
		result.ttype = "login"
	} else {
		if value, ok := record["app"]; ok {
			result.ttype = "app"
			result.appName = value.(string)
		} else {
			result.ttype = "user"
		}
	}

	return result, nil
}
