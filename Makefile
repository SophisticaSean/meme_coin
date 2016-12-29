
containerRunning := $(shell docker ps | grep meme_coin | wc -l)

prebuild: main.go
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o meme_coin .
	docker build -t meme_coin .
	docker rm -f meme_coin; true
console_prebuild: prebuild
	docker run -p 8080:8080 -v ~/mnt/containers/meme_coin:/mnt/containers/meme_coin -v $(shell pwd):/builds/go/src/github.com/SophisticaSean/meme_coin -e console='true' -d --name meme_coin meme_coin
build: prebuild
	docker run -p 8080:8080 -v ~/mnt/containers/meme_coin:/mnt/containers/meme_coin -e bot_token=$(bot_token) -e AdminID=$(AdminID) -d --name meme_coin --restart=always meme_coin
testbuild: prebuild
	docker run -p 8080:8080 -v ~/mnt/containers/meme_coin:/mnt/containers/meme_coin -e pw=$(pw) -e email=$(testEmail) -e AdminID=$(AdminID) -d --name meme_coin --restart=always meme_coin
console: console_prebuild
	sleep 3
	docker exec -it meme_coin /meme_coin
test:
# test depends on already running console_prebuild container
  ifeq ($(containerRunning), 0)
		make console_prebuild
  endif
	@docker exec -it meme_coin bash -c "export GOPATH=/builds/go; export TEST=true; cd /builds/go/src/github.com/SophisticaSean/meme_coin; /usr/local/go/bin/go get; /usr/local/go/bin/go test ./..."
psql:
	docker exec -it meme_coin bash -c 'su postgres -c "psql money"'
dump:
	docker exec -it meme_coin bash -c 'su postgres -c "pg_dump money > /mnt/containers/meme_coin/pg/money.psql.dump"'
	sudo mv ~/mnt/containers/meme_coin/pg/money.psql.dump ./$(shell date +%F_%T)_backup.psql
