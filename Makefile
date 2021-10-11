APP=zc-core
APP_EXECUTABLE="./out/$(APP)"
ALL_PACKAGES=$(shell go list ./... | grep -v "tests" | grep -v "cmd")

PORT?=8080

ENV_LOCAL_TEST=\
		DB_NAME=zurichat \
		CLUSTER_URL=mongodb+srv \
		ENV=local\
		PORT=8080

install:
	@go install ./...
	@echo ""	
	@echo "Run: make help 		to get more info on the available zc_core make commands"
	@echo ""
test.unit:
	@go test ./unit-tests/...

test.integration: 
	@$(ENV_LOCAL_TEST) \
	go test -tags=integration ./tests/... -v -count=1

run:
	@go run main.go

build:
	@go build -o bin/main main.go

install_lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1

lint:
	@golangci-lint run

clean:
	@go clean

fresh:
	@go install github.com/mitranim/gow@latest
	@gow run .
	
help:
	@echo "Usage:"
	@echo "	make [command]"
	@echo ""
	@echo "Available commands:"
	@echo "	help				more info on zc_core make commands"
	@echo "	make				installs all project dependencies"
	@echo "	run				runs the application"
	@echo "	fresh				runs the application in a watch mode"
	@echo "	build				builds the application into /bin folder"
	@echo "	clean				go clean command"
	@echo "	install_lint			runs unit tests"
	@echo "	lint				runs golangci linting"
	@echo "	test.unit			runs unit tests"
	@echo "	test.integration		runs integration tests"
