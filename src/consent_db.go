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
	Type    string `json:"type,omitempty" structs:"type"`
	Token   string `json:"token,omitempty" structs:"token"`
	Brief   string `json:"brief,omitempty" structs:"brief"`
	Message string `json:"message,omitempty" structs:"message"`
	Status  string `json:"status,omitempty" structs:"status"`
}

func (dbobj dbcon) createConsentRecord(userTOKEN string, usertype string, usercode string, brief string, message string, status string) {
	now := int32(time.Now().Unix())
	// brief can not be too long, may be hash it ?
	if len(brief) > 64 {
		return
	}
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
			dbobj.updateRecord2(TblName.Consent, "token", userTOKEN, "brief", brief, &bdoc, nil)
			return
		}
	}
	ev := consentEvent{
		When:    now,
		Who:     usercode,
		Token:   userTOKEN,
		Type:    usertype,
		Brief:   brief,
		Message: message,
		Status:  status,
	}
	// in any case - insert record
	fmt.Printf("insert consent record\n")
	dbobj.createRecord(TblName.Consent, structs.Map(ev))
}

func (dbobj dbcon) cancelConsentRecord(userTOKEN string, brief string) error {
	// brief can not be too long, may be hash it ?
	if len(brief) > 64 {
		return errors.New("Brief value is too long")
	}
	fmt.Printf("%s %s\n", userTOKEN, brief)
	now := int32(time.Now().Unix())
	// update date, status
	bdoc := bson.M{}
	bdoc["when"] = now
	bdoc["status"] = "cancel"
	dbobj.updateRecord2(TblName.Consent, "token", userTOKEN, "brief", brief, &bdoc, nil)
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
