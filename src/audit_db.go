package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	uuid "github.com/hashicorp/go-uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type auditEvent struct {
	When     int32  `json:"when"`
	Who      string `json:"who"`
	Mode     string `json:"mode"`
	Identity string `json:"identity"`
	Record   string `json:"record"`
	App      string `json:"app"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Msg      string `json:"msg"`
	Debug    string `json:"debug"`
	Before   string `json:"before"`
	After    string `json:"after"`
	Atoken   string `json:"atoken"`
}

func audit(title string, record string, mode string, address string) *auditEvent {
	fmt.Printf("/%s : %s\n", title, record)
	return &auditEvent{Title: title, Mode: mode, Who: address, Record: record, Status: "ok", When: int32(time.Now().Unix())}
}

func auditApp(title string, record string, app string, mode string, address string) *auditEvent {
	fmt.Printf("/%s : %s : %s\n", title, app, record)
	return &auditEvent{Title: title, Mode: mode, Who: address, Record: record, Status: "ok", When: int32(time.Now().Unix())}
}

func (event auditEvent) submit(db dbcon) {
	//fmt.Println("submit event to audit!!!!!!!!!!")
	/*
		bdoc, err := bson.Marshal(event)
		if err != nil {
			fmt.Printf("failed to marshal audit event: %s\n", err)
			return
		}
		var bdoc2 bson.M
		err = bson.Unmarshal(bdoc, &bdoc2)
		if err != nil {
			fmt.Printf("failed to marshal audit event2: %s\n", err)
			return
		}*/
	bdoc := bson.M{}
	atoken, _ := uuid.GenerateUUID()
	bdoc["atoken"] = atoken
	bdoc["when"] = event.When
	if len(event.Who) > 0 {
		bdoc["who"] = event.Who
	}
	if len(event.Mode) > 0 {
		bdoc["mode"] = event.Mode
	}
	if len(event.Identity) > 0 {
		bdoc["identity"] = event.Identity
	}
	if len(event.Record) > 0 {
		bdoc["record"] = event.Record
	}
	if len(event.App) > 0 {
		bdoc["app"] = event.App
	}
	if len(event.Title) > 0 {
		bdoc["title"] = event.Title
	}
	bdoc["status"] = event.Status
	if len(event.Msg) > 0 {
		bdoc["msg"] = event.Msg
	}
	if len(event.Debug) > 0 {
		bdoc["debug"] = event.Debug
	}
	if len(event.Before) > 0 {
		bdoc["before"] = event.Before
	}
	if len(event.After) > 0 {
		bdoc["after"] = event.After
	}
	_, err := db.createRecord(TblName.Audit, &bdoc)
	//_, err := db.audit.InsertOne(context.TODO(), &bdoc)
	if err != nil {
		fmt.Printf("failed to marshal audit event: %s\n", err)
		return
	}
	//fmt.Println("AUDIT done!!!")
}

func (dbobj dbcon) getAuditEvents(userTOKEN string, offset int32, limit int32) ([]byte, int64, error) {
	//var results []*auditEvent
	count, err := dbobj.countRecords(TblName.Audit, "record", userTOKEN)
	if err != nil {
		return nil, 0, err
	}
	if count == 0 {
		return []byte("[]"), 0, err
	}
	var results []bson.M
	records, err := dbobj.getList(TblName.Audit, "record", userTOKEN, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	for _, element := range records {
		element["more"] = false
		if _, ok := element["before"]; ok {
			element["more"] = true
			element["before"] = ""
		}
		if _, ok := element["after"]; ok {
			element["more"] = true
			element["after"] = ""
		}
		if _, ok := element["debug"]; ok {
			element["more"] = true
			element["debug"] = ""
		}
		results = append(results, element)
	}

	resultJSON, err := json.Marshal(records)
	if err != nil {
		return nil, 0, err
	}
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, count, nil
}

func (dbobj dbcon) getAuditEvent(atoken string) (string, []byte, error) {
	//var results []*auditEvent
	record, err := dbobj.getRecord(TblName.Audit, "atoken", atoken)
	if err != nil {
		return "", nil, err
	}
	if len(record) == 0 {
		return "", nil, errors.New("not found")
	}
	//fmt.Printf("audit record: %s\n", record)
	before := ""
	after := ""
	debug := ""
	if value, ok := record["before"]; ok {
		before = value.(string)
	}
	if value, ok := record["after"]; ok {
		after = value.(string)
	}
	if value, ok := record["debug"]; ok {
		debug = value.(string)
	}
	//recBson := bson.M{}
	userTOKEN := ""
	if value, ok := record["record"]; ok {
		userTOKEN = value.(string)
		if len(userTOKEN) > 0 {
			if len(before) > 0 {
				before2, after2, _ := dbobj.userDecrypt2(userTOKEN, before, after)
				log.Printf("before: %s", before2)
				log.Printf("after: %s", after2)
				record["before"] = before2
				record["after"] = after2
				if len(debug) == 0 {
					result := fmt.Sprintf(`{"before":%s,"after":%s}`, before2, after2)
					return userTOKEN, []byte(result), nil
				}
				result := fmt.Sprintf(`{"before":%s,"after":%s,"debug":"%s"}`, before2, after2, debug)
				return userTOKEN, []byte(result), nil
			}
			if len(after) > 0 {
				after2, _ := dbobj.userDecrypt(userTOKEN, after)
				log.Printf("after: %s", after2)
				record["after"] = after2
				result := fmt.Sprintf(`{"after":%s,"debug":"%s"}`, after2, debug)
				return userTOKEN, []byte(result), nil
			}
			if len(debug) > 0 {
				result := fmt.Sprintf(`{"debug":"%s"}`, debug)
				return userTOKEN, []byte(result), nil
			}
			return userTOKEN, []byte("{}"), nil
		}
	}
	return userTOKEN, nil, errors.New("not found")
}
