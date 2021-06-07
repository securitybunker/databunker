package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/securitybunker/databunker/src/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type legalBasis struct {
	Brief        string `json:"brief" structs:"brief"`
	Status       string `json:"status" structs:"status"`
	Module       string `json:"module,omitempty" structs:"module,omitempty"`
	Shortdesc    string `json:"shortdesc,omitempty" structs:"shortdesc,omitempty"`
	Fulldesc     string `json:"fulldesc,omitempty" structs:"fulldesc,omitempty"`
	Basistype    string `json:"basistype,omitempty" structs:"basistype"`
	Requiredmsg  string `json:"requiredmsg,omitempty" structs:"requiredmsg,omitempty"`
	Usercontrol  bool   `json:"usercontrol" structs:"usercontrol"`
	Requiredflag bool   `json:"requiredflag" structs:"requiredflag"`
	Creationtime int32  `json:"creationtime" structs:"creationtime"`
}

func (dbobj dbcon) createLegalBasis(brief string, newbrief string, module string, shortdesc string,
	fulldesc string, basistype string, requiredmsg string, status string,
	usercontrol bool, requiredflag bool) (bool, error) {
	bdoc := bson.M{}
	bdoc["basistype"] = basistype
	bdoc["module"] = module
	bdoc["shortdesc"] = shortdesc
	bdoc["fulldesc"] = fulldesc
	if requiredflag == true {
		bdoc["requiredmsg"] = requiredmsg
	} else {
		bdoc["requiredmsg"] = ""
	}
	bdoc["status"] = status
	bdoc["usercontrol"] = usercontrol
	bdoc["requiredflag"] = requiredflag
	raw, err := dbobj.store.GetRecord(storage.TblName.Legalbasis, "brief", brief)
	if err != nil {
		fmt.Printf("error to find:%s", err)
		return false, err
	}
	if raw != nil {
		if len(newbrief) > 0 && newbrief != brief {
			bdoc["brief"] = newbrief
		}
		dbobj.store.UpdateRecord(storage.TblName.Legalbasis, "brief", brief, &bdoc)
		return true, nil
	}
	bdoc["brief"] = brief
	now := int32(time.Now().Unix())
	bdoc["creationtime"] = now
	_, err = dbobj.store.CreateRecord(storage.TblName.Legalbasis, &bdoc)
	if err != nil {
		fmt.Printf("error to insert record: %s\n", err)
		return false, err
	}
	return true, nil
}

func (dbobj dbcon) deleteLegalBasis(brief string) (bool, error) {
	// look up for user with this legal basis
	count, err := dbobj.store.CountRecords(storage.TblName.Agreements, "brief", brief)
	if err != nil {
		return false, err
	}
	if count == 0 {
		// we can safely remove this record
		dbobj.store.DeleteRecord(storage.TblName.Legalbasis, "brief", brief)
		return true, nil
	}
	// change status to revoked for users
	bdoc := bson.M{}
	now := int32(time.Now().Unix())
	bdoc["when"] = now
	bdoc["status"] = "revoked"
	dbobj.store.UpdateRecord2(storage.TblName.Agreements, "brief", brief, "status", "yes", &bdoc, nil)
	bdoc2 := bson.M{}
	bdoc2["status"] = "deleted"
	dbobj.store.UpdateRecord(storage.TblName.Legalbasis, "brief", brief, &bdoc2)
	return true, nil
}

func (dbobj dbcon) revokeLegalBasis(brief string) (bool, error) {
	// look up for user with this legal basis
	bdoc := bson.M{}
	now := int32(time.Now().Unix())
	bdoc["who"] = "admin"
	bdoc["when"] = now
	bdoc["status"] = "revoked"
	dbobj.store.UpdateRecord2(storage.TblName.Agreements, "brief", brief, "status", "yes", &bdoc, nil)
	return true, nil
}

func (dbobj dbcon) getLegalBasisCookieConf() ([]byte, []byte, int, error) {
	records, err := dbobj.store.GetList(storage.TblName.Legalbasis, "status", "active", 0, 0, "requiredflag")
	if err != nil {
		return nil, nil, 0, err
	}
	count := len(records)
	if count == 0 {
		return []byte("[]"), []byte("[]"), 0, err
	}
	count = 0
	var results []bson.M
	cookies := make(map[string]bool)
	for _, element := range records {
		if _, ok := element["module"]; ok {
			if element["module"].(string) == "cookie-popup" {
				cookies[element["brief"].(string)] = true
				results = append(results, element)
				count = count + 1
			}
		}
	}
	//fmt.Printf("cookies %v\n", cookies)
	//fmt.Printf("results %v\n", results)
	if count == 0 {
		return []byte("[]"), []byte("[]"), 0, err
	}
	var scripts []bson.M
	records0, err := dbobj.store.GetList0(storage.TblName.Processingactivities, 0, 0, "")
	for _, record := range records0 {
		if record["legalbasis"] != nil && record["script"] != nil && len(record["script"].(string)) > 0 {
			//fmt.Printf("checking processingactivity record %v\n", record)
			var found []string
			briefs := strings.Split(record["legalbasis"].(string), ",")
			if len(briefs) > 0 {
				for _, brief := range briefs {
					if _, ok := cookies[brief]; ok {
						found = append(found, brief)
					}
				}
			}
			if len(found) > 0 {
				bdoc := bson.M{}
				bdoc["script"] = record["script"]
				bdoc["briefs"] = found
				//fmt.Println("appending bdoc script")
				scripts = append(scripts, bdoc)
			}
		}
	}
	resultJSON, err := json.Marshal(results)
	if err != nil {
		return nil, nil, 0, err
	}
	//fmt.Println("going to marshal scripts")
	scriptsJSON, err := json.Marshal(scripts)
	if err != nil {
		return resultJSON, []byte("[]"), 0, err
	}
	return resultJSON, scriptsJSON, count, nil
}

func (dbobj dbcon) getLegalBasisRecords() ([]byte, int, error) {
	records, err := dbobj.store.GetList0(storage.TblName.Legalbasis, 0, 0, "")
	if err != nil {
		return nil, 0, err
	}
	count := len(records)
	if count == 0 {
		return []byte("[]"), 0, err
	}
	resultJSON, err := json.Marshal(records)
	if err != nil {
		return nil, 0, err
	}
	//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	return resultJSON, count, nil
}

func (dbobj dbcon) checkLegalBasis(brief string) (bool, error) {
	count, err := dbobj.store.CountRecords(storage.TblName.Legalbasis, "brief", brief)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (dbobj dbcon) getLegalBasis(brief string) (bson.M, error) {
	row, err := dbobj.store.GetRecord(storage.TblName.Legalbasis, "brief", brief)
	if err != nil {
		fmt.Printf("error to find:%s", err)
		return nil, err
	}
	return row, err
}
