package main

import (
	"errors"
	"time"

	uuid "github.com/hashicorp/go-uuid"
	"github.com/securitybunker/databunker/src/storage"
	"go.mongodb.org/mongo-driver/bson"
)

var rootXTOKEN string

func (dbobj dbcon) getRootXtoken() (string, error) {
	record, err := dbobj.store.GetRecord2(storage.TblName.Xtokens, "token", "", "type", "root")
	if record == nil || err != nil {
		return "", err
	}
	return record["xtoken"].(string), nil
}

func (dbobj dbcon) createRootXtoken(customRootXtoken string) (string, error) {
	rootToken, err := dbobj.getRootXtoken()
	if err != nil {
		return "", err
	}
	if len(rootToken) > 0 {
		return "already-initialized", nil
	}
	if len(customRootXtoken) > 0 {
		if customRootXtoken != "DEMO" && !isValidUUID(customRootXtoken) {
			return "", errors.New("bad root token format")
		}
		rootToken = customRootXtoken
	} else {
		rootToken, err = uuid.GenerateUUID()
		if err != nil {
			return "", err
		}
	}
	bdoc := bson.M{}
	bdoc["xtoken"] = hashString(dbobj.hash, rootToken)
	bdoc["type"] = "root"
	bdoc["token"] = ""
	_, err = dbobj.store.CreateRecord(storage.TblName.Xtokens, &bdoc)
	if err != nil {
		return rootToken, err
	}
	return rootToken, nil
}

func (dbobj dbcon) generateUserLoginXtoken(userTOKEN string) (string, string, error) {
	// check if user record exists
	record, err := dbobj.lookupUserRecord(userTOKEN)
	if record == nil || err != nil {
		// not found
		return "", "", errors.New("not found")
	}
	tokenUUID, err := uuid.GenerateUUID()
	if err != nil {
		return "", "", err
	}
	hashedToken := hashString(dbobj.hash, tokenUUID)
	// by default login token for 30 minutes only
	expired := int32(time.Now().Unix()) + 10*60
	bdoc := bson.M{}
	bdoc["token"] = userTOKEN
	bdoc["xtoken"] = hashedToken
	bdoc["type"] = "login"
	bdoc["endtime"] = expired
	_, err = dbobj.store.CreateRecord(storage.TblName.Xtokens, &bdoc)
	return tokenUUID, hashedToken, err
}

func (dbobj dbcon) checkUserAuthXToken(xtokenUUID string) (tokenAuthResult, error) {
	result := tokenAuthResult{}
	if xtokenUUID != "DEMO" && isValidUUID(xtokenUUID) == false {
		return result, errors.New("failed to authenticate")
	}
	xtokenHashed := hashString(dbobj.hash, xtokenUUID)
	if len(rootXTOKEN) > 0 && rootXTOKEN == xtokenHashed {
		//log.Println("It is a root token")
		result.ttype = "root"
		result.name = "root"
		return result, nil
	}
	record, err := dbobj.store.GetRecord(storage.TblName.Xtokens, "xtoken", xtokenHashed)
	if record == nil || err != nil {
		return result, errors.New("failed to authenticate")
	}
	tokenType := record["type"].(string)
	//log.Printf("xtoken type: %s\n", tokenType)
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
