.SILENT:

build:
	go build -o ./.bin/bot ./cmd/bot/main.go

run: build
	./.bin/bot

clean:
	rm -rf ./.bin
