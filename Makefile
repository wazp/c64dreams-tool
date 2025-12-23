APP := c64dreams-tool
CMD := ./cmd/c64dreams-tool
DIST := dist

.PHONY: build clean

build:
	go build -o $(DIST)/$(APP) $(CMD)

build-all:
	GOOS=linux   GOARCH=amd64 go build -o $(DIST)/$(APP)-linux-amd64 $(CMD)
	GOOS=linux   GOARCH=arm64 go build -o $(DIST)/$(APP)-linux-arm64 $(CMD)
	GOOS=darwin  GOARCH=arm64 go build -o $(DIST)/$(APP)-macos-arm64 $(CMD)
	GOOS=windows GOARCH=amd64 go build -o $(DIST)/$(APP)-windows-amd64.exe $(CMD)

clean:
	rm -rf $(DIST)
