module github.com/securitybunker/databunker/src

go 1.13

replace github.com/securitybunker/databunker/src/storage => ./storage

require (
	github.com/afocus/captcha v0.0.0-20191010092841-4bd1f21c8868
	github.com/evanphx/json-patch v5.7.0+incompatible
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/gobuffalo/envy v1.10.2 // indirect
	github.com/gobuffalo/packd v1.0.2 // indirect
	github.com/gobuffalo/packr v1.30.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/julienschmidt/httprouter v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-sqlite3 v1.14.19 // indirect
	github.com/oschwald/geoip2-golang v1.9.0
	github.com/oschwald/maxminddb-golang v1.12.0 // indirect
	github.com/prometheus/client_golang v1.17.0
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/qri-io/jsonpointer v0.1.1
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/schollz/sqlite3dump v1.3.1 // indirect
	github.com/securitybunker/databunker/src/storage v0.0.0
	github.com/securitybunker/jsonschema v0.2.1-0.20201128224651-d77c1a3cb787
	github.com/tidwall/gjson v1.17.0
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.2.1
	go.mongodb.org/mongo-driver v1.13.1
	golang.org/x/image v0.14.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sys v0.15.0
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
)
