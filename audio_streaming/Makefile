.PHONY: all
all: server client server_mac client_mac server.exe client.exe zip

server: cmd/server/*go
	GOOS=linux GOARCH=amd64 go build -o server cmd/server/*go

client: cmd/client/*go 
	GOOS=linux GOARCH=amd64 go build -o client cmd/client/*go

server_mac: cmd/server/*go
	GOOS=darwin GOARCH=amd64 go build -o server_mac cmd/server/*go

client_mac: cmd/client/*go
	GOOS=darwin GOARCH=amd64 go build -o client_mac cmd/client/*go

server.exe: cmd/server/*go 
	GOOS=windows GOARCH=amd64 go build -o server.exe cmd/server/*go

client.exe: cmd/client/*go 
	GOOS=windows GOARCH=amd64 go build -o client.exe cmd/client/*go

zip: server client server_mac client_mac server.exe client.exe static/*css static/*html static/*js static/*ico static/images/*png static/audioworklet/* README.md
	zip -q audio_streaming.zip server client server_mac client_mac server.exe client.exe static/*css static/*html static/*js static/*ico static/images/*png static/audioworklet/* README.md
	rm server client server_mac client_mac server.exe client.exe
	

clean:
	rm -f server client server_mac client_mac server.exe client.exe audio_streaming.zip
