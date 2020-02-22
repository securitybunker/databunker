# build without debug
go build -ldflags "-w" -o databunker ./src/bunker.go ./src/xtokens_db.go \
  ./src/utils.go ./src/cryptor.go ./src/notify.go \
  ./src/audit_db.go ./src/audit_api.go \
  ./src/sms.go ./src/email.go \
  ./src/requests_db.go ./src/requests_api.go \
  ./src/users_db.go ./src/users_api.go \
  ./src/userapps_db.go ./src/userapps_api.go \
  ./src/sessions_db.go ./src/sessions_api.go \
  ./src/consent_db.go ./src/consent_api.go \
  ./src/sharedrecords_db.go ./src/sharedrecords_api.go

