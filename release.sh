# build without debug
packr
go build -ldflags "-w" -o databunker ./src/bunker.go ./src/qldb.go ./src/audit_db.go ./src/audit_api.go \
  ./src/utils.go ./src/cryptor.go \
  ./src/sms.go ./src/email.go \
  ./src/users_db.go ./src/users_api.go \
  ./src/userapps_db.go ./src/userapps_api.go \
  ./src/sessions_db.go ./src/sessions_api.go \
  ./src/consent_db.go ./src/consent_api.go \
  ./src/xtokens_db.go ./src/xtokens_api.go ./src/a_main-packr.go
packr clean