package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/paranoidguy/databunker/src/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type legalBasis struct {
	Brief         string `json:"brief" structs:"brief"`
	Status        string `json:"status" structs:"status"`
	Module        string `json:"module,omitempty" structs:"module,omitempty"`
	Shortdesc     string `json:"shortdesc,omitempty" structs:"shortdesc,omitempty"`
	Fulldesc      string `json:"fulldesc,omitempty" structs:"fulldesc,omitempty"`
	Basistype     string `json:"basistype,omitempty" structs:"basistype"`
	Requiredmsg   string `json:"requiredmsg,omitempty" structs:"requiredmsg,omitempty"`
	Usercontrol   bool   `json:"usercontrol" structs:"usercontrol"`
	Requiredflag  bool   `json:"requiredflag" structs:"requiredflag"`
	Creationtime  int32  `json:"creationtime" structs:"creationtime"`
}

func (dbobj dbcon) createLegalBasis(brief string, newbrief string, module string, shortdesc string, fulldesc string, basistype string, requiredmsg string,
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
	bdoc["status"] = "active";
	bdoc["usercontrol"] = usercontrol
	bdoc["requiredflag"] = requiredflag
	raw, err := dbobj.store.GetRecord(storage.TblName.Legalbasis, "brief", brief)
	if err != nil {
		fmt.Printf("error to find:%s", err)
		return false, err
	}
	if raw != nil {
		if len(newbrief) > 0 && newbrief != brief {
			bdoc["brief"] = newbrief;
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