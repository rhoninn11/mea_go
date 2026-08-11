go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1

ls $(go env GOPATH)/bin/sqlc

go mod init szkolaniczego.xd/mym

sqlc generate

go get modernc.org/sqlite

