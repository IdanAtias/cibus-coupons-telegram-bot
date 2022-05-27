build:
	go build -o ./cmd/bot ./cmd
build_linux:
	GOOS=linux GOARCH=amd64 go build -o ./cmd/bot-linux ./cmd
build_zip: build_linux
	zip ./cmd/bot-linux.zip ./cmd/bot-linux
