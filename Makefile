.PHONY: dev generate opendb
	
build:
	make generate
	go build -o chatty cmd/main.go	

dev:
	air

generate:
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/*.proto

opendb:
	sqlite3 test.db