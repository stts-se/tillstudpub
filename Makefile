.PHONY: all
all: init audio_streaming zip

init:
	mkdir -p dist

audio_streaming: init
	cd audio_streaming; \
	GOOS=linux GOARCH=amd64 go build -o ../dist/audio_streaming/audstr_server cmd/audstr_server/*go; \
	GOOS=linux GOARCH=amd64 go build -o ../dist/audio_streaming/audstr_client cmd/audstr_client/*go; \
	GOOS=windows GOARCH=amd64 go build -o ../dist/audio_streaming/audstr_server.exe cmd/audstr_server/*go; \
	GOOS=windows GOARCH=amd64 go build -o ../dist/audio_streaming/audstr_client.exe cmd/audstr_client/*go; \
	cp -r static ../dist/audio_streaming

promptrec: init
	cd promptrec; \
	GOOS=linux GOARCH=amd64 go build -o ../dist/promptrec/promptrec_server *go; \
	GOOS=windows GOARCH=amd64 go build -o ../dist/promptrec/promptrec_server.exe *go
	mkdir -p dist/promptrec/projects/demo-blommor
	cp promptrec/projects/demo-blommor/text.txt dist/promptrec/projects/demo-blommor
	cp -r promptrec/static dist/promptrec

webrtc: init
	cd webrtc_demo; \
	GOOS=linux GOARCH=amd64 go build -o ../dist/webrtc_demo/webrtc_demo *go; \
	GOOS=windows GOARCH=amd64 go build -o ../dist/webrtc_demo/webrtc_demo.exe *go; \
	cp -r index.html README.md ../dist/webrtc_demo


zip: init audio_streaming promptrec webrtc
	mkdir -p dist
	cp README.md dist
	cd dist; zip -q -r audio_demo.zip audio_streaming promptrec webrtc_demo README.md; \
	rm -fr audio_streaming webrtc_demo promptrec README.md
	@echo "Output build saved in dist/audio_demo.zip"

clean:
	rm -rf dist
