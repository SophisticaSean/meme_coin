build: main.go
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o meme_coin .
	docker build -t meme_coin .
	docker rm -f meme_coin; true
	docker run -v ~/mnt/containers/meme_coin:/mnt/containers/meme_coin -e pw=$(pw) -e email=$(email) -d --name meme_coin meme_coin
