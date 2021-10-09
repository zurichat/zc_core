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
		go install ./...	

test.unit:
		go test ./unit-tests/...

test.integration: 
		$(ENV_LOCAL_TEST) \
		go test -tags=integration ./tests/... -v -count=1
