SHELL = /bin/bash
CMDPATH = $(shell ls -1 cmd/*/main.go | head -1)

export GO111MODULE=on

clean:
	@rm -f report.json
	@rm -f coverage.out
	@rm -rf bin

format:
	go fmt ./internal/... ./

install:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.21.0
	go mod download

test:
	go test -cover -v ./internal/... ./

.PHONY: test-report
test-report: 
	go test -cover -v -coverprofile=coverage.out -json ./internal/... ./ | tee report.json

build: clean install
	go mod vendor
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -a \
		-installsuffix cgo \
		-ldflags='-w -s $(BUILD_OVERRIDES)' \
		-o main $(CMDPATH)	

lint:
	golangci-lint run
