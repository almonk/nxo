binary = nxo

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/$(binary)_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build -o ./bin/$(binary)_darwin_arm64
