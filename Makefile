build:
	go build -o dist/roller

build-darwin:
	env GOOS=darwin GOARCH=amd64 go build -o dist/darwin/roller

build-linux:
	env GOOS=linux GOARCH=amd64 go build -o dist/linux/roller

