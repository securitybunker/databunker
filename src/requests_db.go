package main

import (
	"time"

	uuid "github.com/hashicorp/go-uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type requestEvent struct {
	// time for update?
	Creationtime int32  `json:"creationtime"`
	When         int32  `json:"when"`
	Token        string `json:"token"`
	App          string `json:"app,omitempty"`
	Action       string `json:"action"`
	Status       string `json:"status"`
	Change       string `json:"change,omitempty"`
	Rtoken       string `json:"rtoken"`
}

func (dbobj dbcon) saveUserRequest(action string, token string, app string, change string) (string, error) {
	now := int32(time.Now().Unix())
	bdoc := bson.M{}
	rtoken, _ := uuid.GenerateUUID()
	bdoc["when"] = now
	bdoc["token"] = token
	bdoc["action"] = action
	bdoc["rtoken"] = rtoken
	bdoc["creationtime"] = now
	bdoc["status"] = "open"
	if len(change) > 0 {
		bdoc["change"] = change
	}
	if len(app) > 0 {
		bdoc["app"] = app
	}
	_, err := dbobj.createRecord(TblName.Requests, &bdoc)
	return rtoken, err
}
