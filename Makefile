containerRunning := $(shell docker ps | grep meme_coin | wc -l)

all: docker_build
meme_coin: $(wildcard */*.go)
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o meme_coin .
docker_build: meme_coin Dockerfile docker_start.sh
	docker build -t meme_coin .
run: docker_build
	docker run -p 8080:8080 -v ~/mnt/containers/meme_coin:/mnt/containers/meme_coin -e bot_token=$(bot_token) -e AdminID=$(AdminID) -d --name meme_coin --restart=always meme_coin
run_test: docker_build
	docker run -p 8080:8080 -v ~/mnt/containers/meme_coin:/mnt/containers/meme_coin -e pw=$(pw) -e email=$(testEmail) -e AdminID=$(AdminID) -d --name meme_coin --restart=always meme_coin
run_console: docker_build
	docker run -p 8080:8080 -v ~/mnt/containers/meme_coin:/mnt/containers/meme_coin -v $(shell pwd):/builds/go/src/github.com/SophisticaSean/meme_coin -e console='true' -d --name meme_coin meme_coin
console: run_console
	sleep 3
	docker exec -it meme_coin /meme_coin
test: docker_build
  ifeq ($(containerRunning), 0)
		$(MAKE) run_console
  endif
	docker exec -it meme_coin bash -c "export GOPATH=/builds/go; export TEST=true; cd /builds/go/src/github.com/SophisticaSean/meme_coin; /usr/local/go/bin/go get; /usr/local/go/bin/go test ./..."
psql:
	docker exec -it meme_coin bash -c 'su postgres -c "psql money"'
dump:
	docker exec -it meme_coin bash -c 'su postgres -c "pg_dump money > /mnt/containers/meme_coin/pg/money.psql.dump"'
	sudo mv ~/mnt/containers/meme_coin/pg/money.psql.dump ./$(shell date +%F_%T)_backup.psql
clean:
	docker rm -f meme_coin; true
	rm meme_coin

.PHONY: all docker_build run run_test run_console console test psql dump clean