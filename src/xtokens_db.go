package main

import (
	"errors"
	"fmt"
	"time"

	uuid "github.com/hashicorp/go-uuid"
	"go.mongodb.org/mongo-driver/bson"
)

var rootXTOKEN string

func (dbobj dbcon) getRootXtoken() (string, error) {
	record, err := dbobj.getRecord(TblName.Xtokens, "type", "root")
	if err != nil {
		return "", err
	}
	if record == nil {
		return "", nil
	}
	return record["xtoken"].(string), nil
}

func (dbobj dbcon) createRootXtoken() (string, error) {
	rootToken, err := dbobj.getRootXtoken()
	if err != nil {
		return "", err
	}
	if len(rootToken) > 0 {
		return rootToken, nil
	}
	rootToken, err = uuid.GenerateUUID()
	if err != nil {
		return "", err
	}
	bdoc := bson.M{}
	bdoc["xtoken"] = hashString(dbobj.hash, rootToken)
	bdoc["type"] = "root"
	_, err = dbobj.createRecord(TblName.Xtokens, bdoc)
	if err != nil {
		return rootToken, err
	}
	return rootToken, nil
}

func (dbobj dbcon) generateUserLoginXtoken(userTOKEN string) (string, error) {
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
	bdoc["xtoken"] = hashString(dbobj.hash, tokenUUID)
	bdoc["type"] = "login"
	bdoc["endtime"] = expired
	_, err = dbobj.createRecord(TblName.Xtokens, bdoc)
	if err != nil {
		return "", err
	}
	return tokenUUID, nil
}

/*
func (dbobj dbcon) checkXtoken(xtokenUUID string) bool {
	//fmt.Printf("Token0 %s\n", tokenUUID)
	if isValidUUID(xtokenUUID) == false {
		return false
	}
	xtokenHashed := hashString(dbobj.hash, xtokenUUID)
	if len(rootXTOKEN) > 0 && rootXTOKEN == xtokenHashed {
		fmt.Println("It is a root token")
		return true
	}
	record, err := dbobj.getRecord(TblName.Xtokens, "xtoken", xtokenHashed)
	if record == nil || err != nil {
		return false
	}
	tokenType := record["type"].(string)
	if tokenType == "root" {
		rootXTOKEN = xtokenHashed
		return true
	}
	return false
}
*/

func (dbobj dbcon) checkUserAuthXToken(xtokenUUID string) (tokenAuthResult, error) {
	result := tokenAuthResult{}
	if isValidUUID(xtokenUUID) == false {
		return result, errors.New("failed to authenticate")
	}
	xtokenHashed := hashString(dbobj.hash, xtokenUUID)
	if len(rootXTOKEN) > 0 && rootXTOKEN == xtokenHashed {
		//fmt.Println("It is a root token")
		result.ttype = "root"
		result.name = "root"
		return result, nil
	}
	record, err := dbobj.getRecord(TblName.Xtokens, "xtoken", xtokenHashed)
	if record == nil || err != nil {
		return result, errors.New("failed to authenticate")
	}
	tokenType := record["type"].(string)
	fmt.Printf("token type: %s\n", tokenType)
	if tokenType == "root" {
		// we have this admin user
		rootXTOKEN = xtokenHashed
		result.ttype = "root"
		result.name = "root"
		return result, nil
	}
	result.name = xtokenHashed
	// tokenType = temp
	now := int32(time.Now().Unix())
	if now > record["endtime"].(int32) {
		return result, errors.New("xtoken expired")
	}
	result.token = record["token"].(string)
	result.ttype = tokenType
	return result, nil
}
