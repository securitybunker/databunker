module github.com/securitybunker/databunker/src/utils

go 1.21

toolchain go1.23.2

replace github.com/securitybunker/databunker/src/audit => ../audit

require (
	github.com/securitybunker/databunker/src/audit v0.0.0
	github.com/ttacon/libphonenumber v1.2.1
	golang.org/x/sys v0.28.0
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	google.golang.org/protobuf v1.36.1 // indirect
)
