package main

import (
	"encoding/json"
	"fmt"
	"time"

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
	Meta     string `json:"meta"`
}

func audit(title string, record string, mode string, address string) *auditEvent {
	fmt.Printf("/%s : %s\n", title, record)
	return &auditEvent{Title: title, Mode: mode, Who: address, Record: record, Status: "ok", When: int32(time.Now().Unix())}
}

func auditApp(title string, record string, app string) *auditEvent {
	fmt.Printf("/%s : %s : %s\n", title, app, record)
	return &auditEvent{Title: title, Record: record, Status: "ok", When: int32(time.Now().Unix())}
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
	bdoc["when"] = event.When
	if len(event.Who) > 0 {
		bdoc["who"] = event.Who
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
	if len(event.Meta) > 0 {
		bdoc["meta"] = event.Meta
	}
	_, err := db.createRecord(TblName.Audit, &bdoc)
	//_, err := db.audit.InsertOne(context.TODO(), &bdoc)
	if err != nil {
		fmt.Printf("failed to marshal audit event: %s\n", err)
		return
	}
	//fmt.Printf("done!!!")
}

func (dbobj dbcon) getAuditEvents(userTOKEN string, offset int32, limit int32) ([]byte, int64, error) {
	//var results []*auditEvent
	count, err := dbobj.countRecords(TblName.Audit, "record", userTOKEN)
	if err != nil {
		return nil, 0, err
	}
	records, err := dbobj.getList(TblName.Audit, "record", userTOKEN, offset, limit)
	resultJSON, err := json.Marshal(records)
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, count, nil
}
