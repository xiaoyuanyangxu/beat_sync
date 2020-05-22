all: macos windows

macos:
	go build -o beat_sync main.go

windows:
	GOOS=windows GOARCH=386 go build -o beet_sync_386.exe main.go
	GOOS=windows GOARCH=amd64 go build -o beet_sync_amd64.exe main.go
	
