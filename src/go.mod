module github.com/securitybunker/databunker/src

go 1.26.0

// replace github.com/securitybunker/databunker/src/storage => ./storage
// replace github.com/securitybunker/databunker/src/utils => ./utils

require (
	github.com/afocus/captcha v0.0.0-20261010084330-e33d8d71d730
	github.com/evanphx/json-patch v5.9.11+incompatible
	github.com/gobuffalo/packr v1.30.1
	github.com/hashicorp/go-uuid v1.0.4
	github.com/julienschmidt/httprouter v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/oschwald/geoip2-golang v1.13.0
	github.com/prometheus/client_golang v1.25.0
	github.com/qri-io/jsonpointer v0.1.1
	//      github.com/securitybunker/databunker/src/storage v0.0.0
	//      github.com/securitybunker/databunker/src/utils v0.0.0
	github.com/securitybunker/jsonschema v0.2.1-0.20201128224651-d77c1a3cb787
	github.com/tidwall/gjson v1.20.0
	go.mongodb.org/mongo-driver v1.17.10
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/securitybunker/databunker/src/storage v0.0.0-20260728093649-1d96827d1cd6
	github.com/securitybunker/databunker/src/utils v0.0.0-20260728093649-1d96827d1cd6
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dustin/go-humanize v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.10.1 // indirect
	github.com/gobuffalo/envy v1.10.2 // indirect
	github.com/gobuffalo/packd v1.0.2 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/lib/pq v1.12.3 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/ncruces/go-strftime v1.1.0 // indirect
	github.com/oschwald/maxminddb-golang v1.13.1 // indirect
	github.com/pandatix/go-abnf v0.5.3 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.73.0 // indirect
	github.com/prometheus/procfs v0.23.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rogpeppe/go-internal v1.16.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.2 // indirect
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.2.1 // indirect
	golang.org/x/image v0.47.0 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	modernc.org/libc v1.79.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
	modernc.org/sqlite v1.61.0 // indirect
)
