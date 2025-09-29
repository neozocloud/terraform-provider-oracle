default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	TF_ACC=true ORACLE_HOST=localhost ORACLE_PORT=1521 ORACLE_USERNAME=system ORACLE_PASSWORD=MyPassword123 ORACLE_SERVICE=orclpdb1  go test -v ./... -count=1

testacc:
	ORACLE_HOST=localhost ORACLE_PORT=1521 ORACLE_USERNAME=system ORACLE_PASSWORD=MyPassword123 ORACLE_SERVICE=orclpdb1 TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: fmt lint test testacc build install generate
