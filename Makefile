BINDIR=${CURDIR}/bin

####DATABASE######
install-goose:
	GOBIN=$(BINDIR) go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

run-migrations:
	$(BINDIR)/goose -dir migrations postgres "postgresql://user:password@localhost:5432/loadbalancer?sslmode=disable" up


run: run-migrations
	CONFIG_FILE=./configs/values.yaml go run ./cmd/main/

run-server-1:
	COUNTER=1 PORT=8081 go run ./cmd/server/

run-server-2:
	COUNTER=2 PORT=8082 go run ./cmd/server/

run-server-3:
	COUNTER=3 PORT=8083 go run ./cmd/server/

run-server-4:
	COUNTER=4 PORT=8084 go run ./cmd/server/

