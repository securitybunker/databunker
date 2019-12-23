package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fatih/structs"
	"go.mongodb.org/mongo-driver/bson"
)

type consentEvent struct {
	When    int32  `json:"when,omitempty" structs:"when"`
	Who     string `json:"who,omitempty" structs:"who"`
	Mode    string `json:"mode,omitempty" structs:"mode"`
	Token   string `json:"token" structs:"token"`
	Brief   string `json:"brief,omitempty" structs:"brief"`
	Message string `json:"message,omitempty" structs:"message,omitempty"`
	Status  string `json:"status,omitempty" structs:"status"`
	Endtime int32  `json:"endtime,omitempty" structs:"endtime"`
}

func (dbobj dbcon) createConsentRecord(userTOKEN string, mode string, usercode string, brief string, message string, status string, endtime int32) {
	now := int32(time.Now().Unix())
	if len(userTOKEN) > 0 {
		// first check if this consent exists, then update
		raw, err := dbobj.getRecord2(TblName.Consent, "token", userTOKEN, "brief", brief)
		if err != nil {
			fmt.Printf("error to find:%s", err)
			return
		}
		if raw != nil {
			fmt.Println("update rec")
			// update date, status
			bdoc := bson.M{}
			bdoc["when"] = now
			bdoc["status"] = status
			bdoc["endtime"] = endtime
			dbobj.updateRecord2(TblName.Consent, "token", userTOKEN, "brief", brief, &bdoc, nil)
			return
		}
	} else {
		raw, err := dbobj.getRecord2(TblName.Consent, "who", usercode, "brief", brief)
		if err != nil {
			fmt.Printf("error to find:%s", err)
			return
		}
		if raw != nil {
			fmt.Println("update rec")
			// update date, status
			bdoc := bson.M{}
			bdoc["when"] = now
			bdoc["status"] = status
			bdoc["endtime"] = endtime
			dbobj.updateRecord2(TblName.Consent, "who", usercode, "brief", brief, &bdoc, nil)
			return
		}
	}
	ev := consentEvent{
		When:    now,
		Who:     usercode,
		Token:   userTOKEN,
		Mode:    mode,
		Brief:   brief,
		Message: message,
		Status:  status,
		Endtime: endtime,
	}
	// in any case - insert record
	fmt.Printf("insert consent record\n")
	num, err := dbobj.createRecord(TblName.Consent, structs.Map(ev))
	if err != nil {
		fmt.Printf("error to insert record: %s\n", err)
	} else {
		fmt.Printf("num records created: %d\n", num)
	}
}

// link consent record to userToken
func (dbobj dbcon) linkConsentRecords(userTOKEN string, mode string, usercode string) error {
	bdoc := bson.M{}
	bdoc["token"] = userTOKEN
	_, err := dbobj.updateRecord2(TblName.Consent, "token", "", "who", usercode, &bdoc, nil)
	return err
}

func (dbobj dbcon) cancelConsentRecord(userTOKEN string, brief string, mode string, usercode string) error {
	// brief can not be too long, may be hash it ?
	if len(brief) > 64 {
		return errors.New("Brief value is too long")
	}
	now := int32(time.Now().Unix())
	// update date, status
	bdoc := bson.M{}
	bdoc["when"] = now
	bdoc["mode"] = mode
	bdoc["who"] = usercode
	bdoc["status"] = "cancel"
	if len(userTOKEN) > 0 {
		fmt.Printf("%s %s\n", userTOKEN, brief)
		dbobj.updateRecord2(TblName.Consent, "token", userTOKEN, "brief", brief, &bdoc, nil)
	} else {
		dbobj.updateRecord2(TblName.Consent, "who", usercode, "brief", brief, &bdoc, nil)
	}
	return nil
}

// link consent to user?

func (dbobj dbcon) listConsentRecords(userTOKEN string) ([]byte, int, error) {
	records, err := dbobj.getList(TblName.Consent, "token", userTOKEN, 0, 0)
	if err != nil {
		return nil, 0, err
	}
	count := len(records)
	resultJSON, err := json.Marshal(records)
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, count, nil
}

func (dbobj dbcon) viewConsentRecord(userTOKEN string, brief string) ([]byte, error) {
	record, err := dbobj.getRecord2(TblName.Consent, "token", userTOKEN, "brief", brief)
	if err != nil {
		return nil, err
	}
	resultJSON, err := json.Marshal(record)
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, nil
}

func (dbobj dbcon) filterConsentRecords(brief string, offset int32, limit int32) ([]byte, int64, error) {
	//var results []*auditEvent
	count, err := dbobj.countRecords2(TblName.Consent, "brief", brief, "status", "accept")
	if err != nil {
		return nil, 0, err
	}
	records, err := dbobj.getList(TblName.Consent, "brief", brief, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	// we need to return only list of tokens
	var result []string
	for _, rec := range records {
		result = append(result, rec["token"].(string))
	}
	resultJSON, err := json.Marshal(result)
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, count, nil
}
