package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
	"github.com/securitybunker/databunker/src/storage"
	"github.com/securitybunker/databunker/src/utils"
	"go.mongodb.org/mongo-driver/bson"
)

// oldForm encrypts the way earlier versions did.
func oldForm(t *testing.T, value string) string {
	t.Helper()
	enc, err := utils.BasicStringEncrypt(value, e.db.masterKey, e.db.GetCode())
	if err != nil {
		t.Fatalf("failed to encrypt: %s", err)
	}
	return enc
}

func newForm(t *testing.T, value string) string {
	t.Helper()
	enc, err := utils.BasicStringEncrypt2(value, e.db.masterKey, e.db.GetCode())
	if err != nil {
		t.Fatalf("failed to encrypt: %s", err)
	}
	return enc
}

func addOldAgreement(t *testing.T, mode string, identity string, brief string, status string) {
	t.Helper()
	bdoc := bson.M{"when": int32(1), "creationtime": int32(1), "starttime": int32(0), "endtime": int32(0),
		"brief": brief, "mode": mode, "who": oldForm(t, identity), "token": "", "status": status,
		"agreementmethod": "api", "lastmodifiedby": "admin"}
	if _, err := e.db.store.CreateRecord(storage.TblName.Agreements, &bdoc); err != nil {
		t.Fatalf("failed to add agreement: %s", err)
	}
}

func addOldAudit(t *testing.T, userTOKEN string, when int32, title string) string {
	t.Helper()
	atoken, _ := uuid.GenerateUUID()
	bdoc := bson.M{"atoken": atoken, "when": when, "title": title, "status": "ok",
		"who": oldForm(t, "admin"), "record": oldForm(t, userTOKEN), "debug": "old"}
	if _, err := e.db.store.CreateRecord(storage.TblName.Audit, &bdoc); err != nil {
		t.Fatalf("failed to add audit event: %s", err)
	}
	return atoken
}

// agreementRows returns the stored rows of a brief.
func agreementRows(t *testing.T, brief string) []map[string]interface{} {
	t.Helper()
	rows, err := e.db.store.GetList(storage.TblName.Agreements, "brief", brief, 0, 0, "")
	if err != nil {
		t.Fatalf("failed to list agreements: %s", err)
	}
	return rows
}

func addBrief(t *testing.T, brief string) {
	t.Helper()
	raw, _ := helpCreateLBasis(brief, `{"basistype":"consent","usercontrol":true,"status":"active"}`)
	if raw["status"] != "ok" {
		t.Fatalf("failed to create brief: %v", raw)
	}
}

func TestOldAgreementAccept(t *testing.T) {
	email, brief := "old-accept@example.com", "old-accept"
	addOldAgreement(t, "email", email, brief, "no")
	if changed, err := e.db.acceptAgreement("", "email", email, brief, "yes", "api", "", "admin", 0, 0); err != nil || !changed {
		t.Fatalf("accept: %v %v", changed, err)
	}
	rows := agreementRows(t, brief)
	if len(rows) != 1 {
		t.Fatalf("the old row must be updated, not duplicated: %v", rows)
	}
	if rows[0]["status"] != "yes" || rows[0]["who"] != newForm(t, email) {
		t.Fatalf("the row must be updated and stored in the new form: %v", rows[0])
	}
	// again, now found in the new form
	if changed, err := e.db.acceptAgreement("", "email", email, brief, "yes", "api", "", "admin", 0, 0); err != nil || changed {
		t.Fatalf("second accept must change nothing: %v %v", changed, err)
	}
	if rows = agreementRows(t, brief); len(rows) != 1 {
		t.Fatalf("duplicate row: %v", rows)
	}
}

func TestOldAgreementWithdraw(t *testing.T) {
	email, brief := "old-withdraw@example.com", "old-withdraw"
	addBrief(t, brief)
	addOldAgreement(t, "email", email, brief, "yes")
	addOldAgreement(t, "email", "other-withdraw@example.com", brief, "yes")
	if raw, _ := helpWithdrawAgreement("email", email, brief); raw["status"] != "ok" {
		t.Fatalf("withdraw: %v", raw)
	}
	got := map[interface{}]interface{}{}
	for _, row := range agreementRows(t, brief) {
		got[row["who"]] = row["status"]
	}
	if len(got) != 2 || got[newForm(t, email)] != "no" || got[oldForm(t, "other-withdraw@example.com")] != "yes" {
		t.Fatalf("only the given identity may be withdrawn: %v", got)
	}
}

func TestOldAgreementListByIdentity(t *testing.T) {
	email := "old-list@example.com"
	addOldAgreement(t, "email", email, "old-list1", "yes")
	addOldAgreement(t, "email", "other-list@example.com", "old-list1", "yes")
	if _, err := e.db.acceptAgreement("", "email", email, "old-list2", "yes", "api", "", "admin", 0, 0); err != nil {
		t.Fatalf("accept: %s", err)
	}
	data, count, err := e.db.listAgreementRecordsByIdentity(email)
	if err != nil || count != 2 {
		t.Fatalf("expected both forms, got %d %v %s", count, err, data)
	}
	if strings.Count(string(data), email) != 2 || strings.Contains(string(data), "other-list") {
		t.Fatalf("wrong rows: %s", data)
	}
	if data, count, err = e.db.listAgreementRecordsByIdentity("nobody-list@example.com"); err != nil || count != 0 || string(data) != "[]" {
		t.Fatalf("unknown identity: %d %v %s", count, err, data)
	}
}

func TestOldAgreementLinkedToNewUser(t *testing.T) {
	email, phone := "old-link@example.com", "+14155550123"
	addOldAgreement(t, "email", email, "old-link1", "yes")
	addOldAgreement(t, "phone", phone, "old-link2", "yes")
	addOldAgreement(t, "email", "other-link@example.com", "old-link1", "yes")
	raw, err := helpCreateUser(fmt.Sprintf(`{"login":"oldlink","email":%q,"phone":%q}`, email, phone))
	if err != nil || raw["status"] != "ok" {
		t.Fatalf("failed to create user: %v %v", raw, err)
	}
	userTOKEN := utils.GetStringValue(raw["token"])
	data, count, err := e.db.listAgreementRecords(userTOKEN)
	if err != nil {
		t.Fatalf("list: %s", err)
	}
	var rows []map[string]interface{}
	json.Unmarshal(data, &rows)
	who := map[string]string{}
	for _, row := range rows {
		who[utils.GetStringValue(row["brief"])] = utils.GetStringValue(row["who"])
	}
	if count != 2 || who["old-link1"] != email || who["old-link2"] != phone {
		t.Fatalf("both identities must be linked, got %d: %s", count, data)
	}
	for _, row := range agreementRows(t, "old-link1") {
		if row["who"] == oldForm(t, "other-link@example.com") && utils.GetUuidString(row["token"]) == userTOKEN {
			t.Fatalf("another identity was linked")
		}
	}
}

// The same brief accepted by email and phone leaves one row, as before.
func TestOldAgreementDuplicateBrief(t *testing.T) {
	email, phone := "old-dup@example.com", "+14155550124"
	addOldAgreement(t, "email", email, "old-dup", "yes")
	addOldAgreement(t, "phone", phone, "old-dup", "yes")
	raw, err := helpCreateUser(fmt.Sprintf(`{"login":"olddup","email":%q,"phone":%q}`, email, phone))
	if err != nil || raw["status"] != "ok" {
		t.Fatalf("failed to create user: %v %v", raw, err)
	}
	rows := agreementRows(t, "old-dup")
	if len(rows) != 1 || rows[0]["who"] != newForm(t, email) {
		t.Fatalf("expected only the email row: %v", rows)
	}
}

func TestOldAgreementView(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	addOldAgreement(t, "email", "old-view@example.com", "old-view", "yes")
	bdoc := bson.M{"token": userTOKEN}
	e.db.store.UpdateRecord(storage.TblName.Agreements, "who", oldForm(t, "old-view@example.com"), &bdoc)
	data, err := e.db.viewAgreementRecord(userTOKEN, "old-view")
	if err != nil || !strings.Contains(string(data), "old-view@example.com") {
		t.Fatalf("old identity must be readable: %s %v", data, err)
	}
}

func TestOldAuditPaging(t *testing.T) {
	// 7 old and 5 new rows, newest first: n5..n1, o7..o1
	userTOKEN, _ := uuid.GenerateUUID()
	var want []string
	for i := 1; i <= 7; i++ {
		addOldAudit(t, userTOKEN, int32(100+i), fmt.Sprintf("o%d", i))
		want = append([]string{fmt.Sprintf("o%d", i)}, want...)
	}
	for i := 1; i <= 5; i++ {
		SaveAuditEvent(&AuditEvent{When: int32(200 + i), Who: "admin", Record: userTOKEN, Title: fmt.Sprintf("n%d", i), Status: "ok"}, e.db, Config{})
		want = append([]string{fmt.Sprintf("n%d", i)}, want...)
	}
	for offset := 0; offset <= 14; offset++ {
		for limit := 0; limit <= 14; limit++ {
			rows, total, err := e.db.getAuditRecords(userTOKEN, int32(offset), int32(limit))
			if err != nil || total != 12 {
				t.Fatalf("offset %d limit %d: total %d %v", offset, limit, total, err)
			}
			from := min(offset, 12)
			to := min(from+limit, 12)
			if offset == 0 && limit == 0 {
				to = 12 // no limit
			}
			got := []string{}
			for _, row := range rows {
				got = append(got, utils.GetStringValue(row["title"]))
			}
			if strings.Join(got, ",") != strings.Join(want[from:to], ",") {
				t.Fatalf("offset %d limit %d: got %v, want %v", offset, limit, got, want[from:to])
			}
		}
	}
	if rows, total, err := e.db.getAuditRecords(userTOKEN, -1, -1); err != nil || total != 12 || len(rows) != 12 {
		t.Fatalf("negative values: %d of %d, %v", len(rows), total, err)
	}
	data, total, err := e.db.getAuditEvents(userTOKEN, 0, 20)
	if err != nil || total != 12 || strings.Count(string(data), `"who":"admin"`) != 12 {
		t.Fatalf("who must be readable in both forms: %d %v %s", total, err, data)
	}
}

func TestAuditPagingOverLimit(t *testing.T) {
	userTOKEN, _ := uuid.GenerateUUID()
	for i := 0; i < 70; i++ {
		addOldAudit(t, userTOKEN, int32(100+i), "o")
		SaveAuditEvent(&AuditEvent{When: int32(1000 + i), Record: userTOKEN, Title: "n", Status: "ok"}, e.db, Config{})
	}
	if rows, total, err := e.db.getAuditRecords(userTOKEN, 0, 500); err != nil || total != 140 || len(rows) != 100 {
		t.Fatalf("expected 100 of 140 rows, got %d of %d, %v", len(rows), total, err)
	}
}

func TestOldAuditEventGet(t *testing.T) {
	raw, err := helpCreateUser(`{"login":"oldaudit1"}`)
	if err != nil || raw["status"] != "ok" {
		t.Fatalf("failed to create user: %v %v", raw, err)
	}
	userTOKEN := utils.GetStringValue(raw["token"])
	atoken := addOldAudit(t, userTOKEN, 100, "old")
	if raw, _ = helpGetUserAuditEvent(atoken); raw["status"] != "ok" {
		t.Fatalf("old event must be readable: %v", raw)
	}
	// new rows must use the new form only
	rows, _ := e.db.store.GetList(storage.TblName.Audit, "record", newForm(t, userTOKEN), 0, 0, "")
	if len(rows) == 0 {
		t.Fatalf("new events must be stored in the new form")
	}
	rows, _ = e.db.store.GetList(storage.TblName.Audit, "record", oldForm(t, userTOKEN), 0, 0, "")
	if len(rows) != 1 {
		t.Fatalf("expected only the seeded old event, got %d", len(rows))
	}
}
