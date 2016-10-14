prebuild: main.go
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o meme_coin .
	docker build -t meme_coin .
	docker rm -f meme_coin; true
build: prebuild
	docker run -v ~/mnt/containers/meme_coin:/mnt/containers/meme_coin -e pw=$(pw) -e email=$(email) -d --name meme_coin meme_coin
console: prebuild
	docker run -v ~/mnt/containers/meme_coin:/mnt/containers/meme_coin -e console='true' -d --name meme_coin meme_coin
	sleep 3
	docker exec -it meme_coin /meme_coin
test:
	docker build -t meme_coin .
	docker rm -f meme_coin; true
	docker run -v ~/mnt/containers/meme_coin:/mnt/containers/meme_coin -e TEST='true' -e console='true' --name meme_coin meme_coin
psql:
	docker exec -it meme_coin bash -c 'su postgres -c "psql money"'
