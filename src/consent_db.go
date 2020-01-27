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
	Creationtime   int32  `json:"creationtime" structs:"creationtime"`
	Starttime      int32  `json:"starttime" structs:"starttime"`
	Endtime        int32  `json:"endtime" structs:"endtime"`
	When           int32  `json:"when,omitempty" structs:"when"`
	Who            string `json:"who,omitempty" structs:"who"`
	Mode           string `json:"mode,omitempty" structs:"mode"`
	Token          string `json:"token" structs:"token"`
	Brief          string `json:"brief,omitempty" structs:"brief"`
	Status         string `json:"status,omitempty" structs:"status"`
	Message        string `json:"message,omitempty" structs:"message,omitempty"`
	Freetext       string `json:"freetext,omitempty" structs:"freetext,omitempty"`
	Lawfulbasis    string `json:"lawfulbasis,omitempty" structs:"lawfulbasis"`
	Consentmethod  string `json:"consentmethod,omitempty" structs:"consentmethod"`
	Referencecode  string `json:"referencecode,omitempty" structs:"referencecode,omitempty"`
	Lastmodifiedby string `json:"lastmodifiedby,omitempty" structs:"lastmodifiedby,omitempty"`
}

func (dbobj dbcon) createConsentRecord(userTOKEN string, mode string, usercode string,
	brief string, message string, status string, lawfulbasis string, consentmethod string,
	referencecode string, freetext string, lastmodifiedby string, starttime int32, endtime int32) (bool, error) {
	now := int32(time.Now().Unix())
	bdoc := bson.M{}
	bdoc["when"] = now
	bdoc["status"] = status
	bdoc["endtime"] = endtime
	if len(lawfulbasis) > 0 {
		// in case of update, consent, use new value
		bdoc["lawfulbasis"] = lawfulbasis
	}
	if len(consentmethod) > 0 {
		bdoc["consentmethod"] = consentmethod
	}
	if len(referencecode) > 0 {
		bdoc["referencecode"] = referencecode
	}
	if len(freetext) > 0 {
		bdoc["freetext"] = freetext
	}
	bdoc["lastmodifiedby"] = lastmodifiedby
	if len(userTOKEN) > 0 {
		// first check if this consent exists, then update
		raw, err := dbobj.getRecord2(TblName.Consent, "token", userTOKEN, "brief", brief)
		if err != nil {
			fmt.Printf("error to find:%s", err)
			return false, err
		}
		if raw != nil {
			dbobj.updateRecord2(TblName.Consent, "token", userTOKEN, "brief", brief, &bdoc, nil)
			if status != raw["status"].(string) {
				// status changed
				return true, nil
			}
			return false, nil
		}
	} else {
		raw, err := dbobj.getRecord2(TblName.Consent, "who", usercode, "brief", brief)
		if err != nil {
			fmt.Printf("error to find:%s", err)
			return false, err
		}
		if raw != nil {
			dbobj.updateRecord2(TblName.Consent, "who", usercode, "brief", brief, &bdoc, nil)
			if status != raw["status"].(string) {
				// status changed
				return true, nil
			}
			return false, err
		}
	}
	if len(consentmethod) == 0 {
		consentmethod = "api"
	}
	if len(lawfulbasis) == 0 {
		lawfulbasis = "consent"
	}
	ev := consentEvent{
		Creationtime:   now,
		Endtime:        endtime,
		When:           now,
		Who:            usercode,
		Token:          userTOKEN,
		Mode:           mode,
		Brief:          brief,
		Message:        message,
		Status:         status,
		Freetext:       freetext,
		Lawfulbasis:    lawfulbasis,
		Consentmethod:  consentmethod,
		Referencecode:  referencecode,
		Lastmodifiedby: lastmodifiedby,
	}
	// in any case - insert record
	_, err := dbobj.createRecord(TblName.Consent, structs.Map(ev))
	if err != nil {
		fmt.Printf("error to insert record: %s\n", err)
		return false, err
	}
	return true, nil
}

// link consent record to userToken
func (dbobj dbcon) linkConsentRecords(userTOKEN string, mode string, usercode string) error {
	bdoc := bson.M{}
	bdoc["token"] = userTOKEN
	_, err := dbobj.updateRecord2(TblName.Consent, "token", "", "who", usercode, &bdoc, nil)
	return err
}

func (dbobj dbcon) withdrawConsentRecord(userTOKEN string, brief string, mode string, usercode string, lastmodifiedby string) error {
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
	bdoc["endtime"] = 0
	bdoc["status"] = "no"
	bdoc["lastmodifiedby"] = lastmodifiedby
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
	if err != nil {
		return nil, 0, err
	}
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, count, nil
}

func (dbobj dbcon) viewConsentRecord(userTOKEN string, brief string) ([]byte, error) {
	record, err := dbobj.getRecord2(TblName.Consent, "token", userTOKEN, "brief", brief)
	if err != nil {
		return nil, err
	}
	resultJSON, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, nil
}

func (dbobj dbcon) filterConsentRecords(brief string, offset int32, limit int32) ([]byte, int64, error) {
	//var results []*auditEvent
	count, err := dbobj.countRecords(TblName.Consent, "brief", brief)
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
	if err != nil {
		return nil, 0, err
	}
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, count, nil
}

func (dbobj dbcon) getConsentTypes() ([]byte, int64, error) {
	records, err := dbobj.getUniqueList(TblName.Consent, "brief")
	if err != nil {
		return nil, 0, err
	}
	count:= int64(len(records))
	// we need to return only list of briefs
	var result []string
	for _, rec := range records {
		result = append(result, rec["brief"].(string))
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, 0, err
	}
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, count, nil
}

func (dbobj dbcon) expireConsentRecords(notifyURL string) error {
	records, err := dbobj.getExpiring(TblName.Consent, "status", "yes")
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
		fmt.Printf("This consent record is expired: %s - %s\n", userTOKEN, brief)
		if len(userTOKEN) > 0 {
			fmt.Printf("%s %s\n", userTOKEN, brief)
			dbobj.updateRecord2(TblName.Consent, "token", userTOKEN, "brief", brief, &bdoc, nil)
			notifyConsentChange(notifyURL, brief, "expired", "token", userTOKEN)
		} else {
			usercode := rec["who"].(string)
			dbobj.updateRecord2(TblName.Consent, "who", usercode, "brief", brief, &bdoc, nil)
			notifyConsentChange(notifyURL, brief, "expired", rec["mode"].(string), usercode)
		}

	}
	return nil
}
