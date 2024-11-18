.PHONY: build run dev generate opendb
	
build:
	make generate
	go build -o chatty cmd/main.go	

run:
	make build
	./chatty

dev:
	air

generate:
	protoc -I=proto/ \
		--go_out=. --go-grpc_out=. --go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=. --openapiv2_opt=use_go_templates=true,paths=source_relative \
		proto/*.proto 

opendb:
	sqlite3 test.db
