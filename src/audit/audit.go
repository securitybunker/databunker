package audit

import "time"

type AuditEvent struct {
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

func CreateAuditEvent(title string, record string, mode string, identity string) *AuditEvent {
	//fmt.Printf("/%s : %s\n", title, record)
	return &AuditEvent{Title: title, Mode: mode, Who: identity, Record: record, Status: "ok", When: int32(time.Now().Unix())}
}

func CreateAuditAppEvent(title string, record string, app string, mode string, identity string) *AuditEvent {
	//fmt.Printf("/%s : %s : %s\n", title, app, record)
	return &AuditEvent{Title: title, Mode: mode, Who: identity, Record: record, Status: "ok", When: int32(time.Now().Unix())}
}
