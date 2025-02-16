ccred=$(shell echo -e "\033[0;31m")
ccyellow=$(shell echo -e "\033[0;33m")
ccend=$(shell echo -e "\033[0m")

run:
	docker compose up

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

sqlc:
	sqlc generate;

unit_tests:
	go test -v -race -coverprofile=./tests/coverage.out  ./internal/...
	go tool cover -html=./tests/coverage.out -o ./tests/coverage.html

integration_tests:
	go test -v -race ./tests/integration/...

e2e:
	@echo "use 'sh run_e2e' to run e2e tests with docker-startup"
	@echo "or just ${ccyellow}go test -v -race ./tests/e2e/...${ccend}"
	@echo
	@echo ">> note, that in both variants you should use ${ccred}POSTGRES_TEST_DB_URL=<url>${ccend}, ${ccred}REDIS_TEST_DB_URL${ccend} and ${ccred}STATUS=testing${ccend} flags"
	@echo "for example: ${ccyellow}POSTGRES_TEST_DB_URL=postgres://root:password@localhost:5432/shop?sslmode=disable REDIS_TEST_DB_URL=localhost:6379 STATUS=testing sh run_e2e.sh${ccend}"

lint:
	golangci-lint run

load_test:
	STATUS=loadtest sh run_load_test.sh

.PHONY: new_migration sqlc unit_tests integration_tests e2e lint run load_test
