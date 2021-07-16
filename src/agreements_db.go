package main

import (
	"encoding/json"
	//"errors"
	"fmt"
	"time"

	"github.com/securitybunker/databunker/src/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type agreement struct {
	Creationtime    int32  `json:"creationtime" structs:"creationtime"`
	Starttime       int32  `json:"starttime" structs:"starttime"`
	Endtime         int32  `json:"endtime" structs:"endtime"`
	When            int32  `json:"when,omitempty" structs:"when"`
	Who             string `json:"who,omitempty" structs:"who"`
	Mode            string `json:"mode,omitempty" structs:"mode"`
	Token           string `json:"token" structs:"token"`
	Brief           string `json:"brief,omitempty" structs:"brief"`
	Status          string `json:"status,omitempty" structs:"status"`
	Referencecode   string `json:"referencecode,omitempty" structs:"referencecode,omitempty"`
	Lastmodifiedby  string `json:"lastmodifiedby,omitempty" structs:"lastmodifiedby,omitempty"`
	Agreementmethod string `json:"agreementmethod,omitempty" structs:"agreementmethod"`
}

func (dbobj dbcon) acceptAgreement(userTOKEN string, mode string, identity string, brief string,
	status string, agreementmethod string, referencecode string, lastmodifiedby string,
	starttime int32, endtime int32) (bool, error) {
	now := int32(time.Now().Unix())
	bdoc := bson.M{}
	bdoc["when"] = now
	bdoc["status"] = status
	bdoc["starttime"] = starttime
	bdoc["endtime"] = endtime
	bdoc["lastmodifiedby"] = lastmodifiedby
	if len(referencecode) > 0 {
		bdoc["referencecode"] = referencecode
	}
	encIdentity := ""
	if len(identity) > 0 {
		encIdentity, _ = basicStringEncrypt(identity, dbobj.masterKey, dbobj.GetCode())
	}
	if len(userTOKEN) > 0 {
		// first check if this agreement exists, then update
		raw, err := dbobj.store.GetRecord2(storage.TblName.Agreements, "token", userTOKEN, "brief", brief)
		if err != nil {
			fmt.Printf("error to find:%s", err)
			return false, err
		}
		if raw != nil {
			dbobj.store.UpdateRecord2(storage.TblName.Agreements, "token", userTOKEN, "brief", brief, &bdoc, nil)
			if status != raw["status"].(string) {
				// status changed
				return true, nil
			}
			return false, nil
		}
	} else if len(identity) > 0 {
		raw, err := dbobj.store.GetRecord2(storage.TblName.Agreements, "who", encIdentity, "brief", brief)
		if err != nil {
			fmt.Printf("error to find:%s", err)
			return false, err
		}
		if raw != nil {
			dbobj.store.UpdateRecord2(storage.TblName.Agreements, "who", encIdentity, "brief", brief, &bdoc, nil)
			if status != raw["status"].(string) {
				// status changed
				return true, nil
			}
			return false, err
		}
	}
	bdoc["brief"] = brief
	bdoc["mode"] = mode
	bdoc["who"] = encIdentity
	bdoc["token"] = userTOKEN
	bdoc["creationtime"] = now
	if len(agreementmethod) > 0 {
		bdoc["agreementmethod"] = agreementmethod
	} else {
		bdoc["agreementmethod"] = "api"
	}
	// in any case - insert record
	_, err := dbobj.store.CreateRecord(storage.TblName.Agreements, &bdoc)
	if err != nil {
		fmt.Printf("error to insert record: %s\n", err)
		return false, err
	}
	return true, nil
}

// link consent record to userToken
func (dbobj dbcon) linkAgreementRecords(userTOKEN string, encIdentity string) error {
	bdoc := bson.M{}
	bdoc["token"] = userTOKEN
	_, err := dbobj.store.UpdateRecord2(storage.TblName.Agreements, "token", "", "who", encIdentity, &bdoc, nil)
	return err
}

func (dbobj dbcon) withdrawAgreement(userTOKEN string, brief string, mode string, identity string, lastmodifiedby string) error {
	now := int32(time.Now().Unix())
	// update date, status
	encIdentity := ""
	if len(identity) > 0 {
		encIdentity, _ = basicStringEncrypt(identity, dbobj.masterKey, dbobj.GetCode())
	}
	bdoc := bson.M{}
	bdoc["when"] = now
	bdoc["mode"] = mode
	bdoc["who"] = encIdentity
	bdoc["endtime"] = 0
	bdoc["status"] = "no"
	bdoc["lastmodifiedby"] = lastmodifiedby
	if len(userTOKEN) > 0 {
		fmt.Printf("%s %s\n", userTOKEN, brief)
		dbobj.store.UpdateRecord2(storage.TblName.Agreements, "token", userTOKEN, "brief", brief, &bdoc, nil)
	} else if len(identity) > 0 {
		dbobj.store.UpdateRecord2(storage.TblName.Agreements, "who", encIdentity, "brief", brief, &bdoc, nil)
	}
	return nil
}

func (dbobj dbcon) listAgreementRecords(userTOKEN string) ([]byte, int, error) {
	records, err := dbobj.store.GetList(storage.TblName.Agreements, "token", userTOKEN, 0, 0, "")
	if err != nil {
		return nil, 0, err
	}
	count := len(records)
	if count == 0 {
		return []byte("[]"), 0, err
	}
	for _, rec := range records {
		encIdentity := rec["who"].(string)
		if len(encIdentity) > 0 {
			identity, _ := basicStringDecrypt(encIdentity, dbobj.masterKey, dbobj.GetCode())
			if len(identity) > 0 {
				rec["who"] = identity
			}
		}
	}
	resultJSON, err := json.Marshal(records)
	if err != nil {
		return nil, 0, err
	}
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, count, nil
}

func (dbobj dbcon) listAgreementRecordsByIdentity(identity string) ([]byte, int, error) {
	encIdentity, _ := basicStringEncrypt(identity, dbobj.masterKey, dbobj.GetCode())
	records, err := dbobj.store.GetList(storage.TblName.Agreements, "who", encIdentity, 0, 0, "")
	if err != nil {
		return nil, 0, err
	}
	count := len(records)
	if count == 0 {
		return []byte("[]"), 0, err
	}
	for _, rec := range records {
		rec["who"] = identity
	}
	resultJSON, err := json.Marshal(records)
	if err != nil {
		return nil, 0, err
	}
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, count, nil
}

func (dbobj dbcon) viewAgreementRecord(userTOKEN string, brief string) ([]byte, error) {
	record, err := dbobj.store.GetRecord2(storage.TblName.Agreements, "token", userTOKEN, "brief", brief)
	if record == nil || err != nil {
		return nil, err
	}
	encIdentity := record["who"].(string)
	if len(encIdentity) > 0 {
		identity, _ := basicStringDecrypt(encIdentity, dbobj.masterKey, dbobj.GetCode())
		if len(identity) > 0 {
			record["who"] = identity
		}
	}
	resultJSON, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, nil
}

func (dbobj dbcon) expireAgreementRecords(notifyURL string) error {
	records, err := dbobj.store.GetExpiring(storage.TblName.Agreements, "status", "yes")
	if err != nil {
		return err
	}
	for _, rec := range records {
		now := int32(time.Now().Unix())
		// update date, status
		bdoc := bson.M{}
		bdoc["when"] = now
		bdoc["status"] = "expired"
		userTOKEN := rec["token"].(string)
		brief := rec["brief"].(string)
		fmt.Printf("This Agreements record is expired: %s - %s\n", userTOKEN, brief)
		if len(userTOKEN) > 0 {
			fmt.Printf("%s %s\n", userTOKEN, brief)
			dbobj.store.UpdateRecord2(storage.TblName.Agreements, "token", userTOKEN, "brief", brief, &bdoc, nil)
			notifyConsentChange(notifyURL, brief, "expired", "token", userTOKEN)
		} else {
			encIdentity := rec["who"].(string)
			dbobj.store.UpdateRecord2(storage.TblName.Agreements, "who", encIdentity, "brief", brief, &bdoc, nil)
			identity, _ := basicStringDecrypt(encIdentity, dbobj.masterKey, dbobj.GetCode())
			notifyConsentChange(notifyURL, brief, "expired", rec["mode"].(string), identity)
		}

	}
	return nil
}
