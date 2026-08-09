SERVER_APP_NAME=mvdb-server
CLI_APP_NAME=mvdb-cli

make :
	lsls

build-cli:
	go build -o ${CLI_APP_NAME} cmd/cli/main.go

run-server: build-server
	./${SERVER_APP_NAME}

run-server-with-config: build-server
	CONFIG_FILE_NAME=config.yml ./${SERVER_APP_NAME}

run-cli: build-cli
	./${CLI_APP_NAME} $(ARGS)

run_unit_test:
	go test ./internal/...

run_e2e_test: build-server
	go test ./test/...

run_test_coverage:
	go test ./... -coverprofile=coverage.out