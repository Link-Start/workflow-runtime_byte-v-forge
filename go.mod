module github.com/byte-v-forge/workflow-runtime

go 1.26

toolchain go1.26.3

require google.golang.org/protobuf v1.36.11

require (
	github.com/byte-v-forge/common-lib v0.0.0
	github.com/jackc/pgx/v5 v5.9.2
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace github.com/byte-v-forge/common-lib => ../common-lib
