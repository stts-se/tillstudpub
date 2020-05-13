.PHONY: all
all: webrtc_server webrtc_server_mac webrtc_server.exe zip

webrtc_server: 
	GOOS=linux GOARCH=amd64 go build -o webrtc_server *go

webrtc_server_mac: 
	GOOS=darwin GOARCH=amd64 go build -o webrtc_server_mac *go

webrtc_server.exe: 
	GOOS=windows GOARCH=amd64 go build -o webrtc_server.exe *go

zip: webrtc_server webrtc_server_mac webrtc_server.exe
	zip -q -r webrtc_demo.zip webrtc_server webrtc_server_mac webrtc_server.exe index.html favicon.ico README.md
	rm webrtc_server webrtc_server_mac webrtc_server.exe 

clean:
	rm -f webrtc_server webrtc_server_mac webrtc_server.exe webtrc_demo.zip
