module github.com/securitybunker/databunker/src

go 1.13

replace github.com/securitybunker/databunker/src/storage => ./storage

require (
	github.com/afocus/captcha v0.0.0-20191010092841-4bd1f21c8868
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/fatih/structs v1.1.0
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gobuffalo/packr v1.30.1
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/hashicorp/go-uuid v1.0.2
	github.com/julienschmidt/httprouter v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/prometheus/client_golang v1.4.1
	github.com/qri-io/jsonpointer v0.1.1
	github.com/schollz/sqlite3dump v1.2.4 // indirect
	github.com/securitybunker/databunker/src/storage v0.0.0
	github.com/securitybunker/jsonschema v0.2.1-0.20201128224651-d77c1a3cb787
	github.com/tidwall/gjson v1.8.1
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.2.1
	go.mongodb.org/mongo-driver v1.7.1
	golang.org/x/image v0.0.0-20210504121937-7319ad40d33e // indirect
	golang.org/x/sys v0.0.0-20200219091948-cb0a6d8edb6c
	gopkg.in/yaml.v2 v2.2.8
)
