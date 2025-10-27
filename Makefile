.SILENT:

build:
	go build -o ./.bin/bot ./cmd/bot/main.go

run: build
	./.bin/bot

clean:
	rm -rf ./.bin

help:
	echo "Available targets:"
	echo "	make build   - Compile the Go application and output binary to ./.bin/"
	echo "	make run     - Build and then run the Go application"
	echo "	make clean   - Remove compiled binaries and build artifacts"