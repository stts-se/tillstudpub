.PHONY: all
all: promptrec_server promptrec_server_mac promptrec_server.exe zip

promptrec_server: 
	GOOS=linux GOARCH=amd64 go build -o promptrec_server *go

promptrec_server_mac: 
	GOOS=darwin GOARCH=amd64 go build -o promptrec_server_mac *go

promptrec_server.exe: 
	GOOS=windows GOARCH=amd64 go build -o promptrec_server.exe *go

zip: promptrec_server promptrec_server_mac promptrec_server.exe
	zip -q -r promptrec_server.zip promptrec_server promptrec_server_mac promptrec_server.exe static projects/demo-blommor/text.txt projects/demo-djur/text.txt favicon.ico README.md
	rm promptrec_server promptrec_server_mac promptrec_server.exe 

clean:
	rm -f promptrec_server promptrec_server_mac promptrec_server.exe webtrc_server.zip
