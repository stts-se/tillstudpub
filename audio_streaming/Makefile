.PHONY: all
all: audstr_server audstr_client audstr_server_mac audstr_client_mac audstr_server.exe audstr_client.exe zip

audstr_server: cmd/audstr_server/*go
	GOOS=linux GOARCH=amd64 go build -o audstr_server cmd/audstr_server/*go

audstr_client: cmd/audstr_client/*go 
	GOOS=linux GOARCH=amd64 go build -o audstr_client cmd/audstr_client/*go

audstr_server_mac: cmd/audstr_server/*go
	GOOS=darwin GOARCH=amd64 go build -o audstr_server_mac cmd/audstr_server/*go

audstr_client_mac: cmd/audstr_client/*go
	GOOS=darwin GOARCH=amd64 go build -o audstr_client_mac cmd/audstr_client/*go

audstr_server.exe: cmd/audstr_server/*go 
	GOOS=windows GOARCH=amd64 go build -o audstr_server.exe cmd/audstr_server/*go

audstr_client.exe: cmd/audstr_client/*go 
	GOOS=windows GOARCH=amd64 go build -o audstr_client.exe cmd/audstr_client/*go

zip: audstr_server audstr_client audstr_server_mac audstr_client_mac audstr_server.exe audstr_client.exe static  README.md
	zip -q -r audio_streaming.zip audstr_server audstr_client audstr_server_mac audstr_client_mac audstr_server.exe audstr_client.exe static README.md
	rm audstr_server audstr_client audstr_server_mac audstr_client_mac audstr_server.exe audstr_client.exe

clean:
	rm -f audstr_server audstr_client audstr_server_mac audstr_client_mac audstr_server.exe audstr_client.exe audio_streaming.zip
