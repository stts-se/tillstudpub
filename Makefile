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
	echo -n "" >| dist/promptrec/projects/demo-blommor/text.txt
	echo "01	lilja	 Säg en första blomma" >> dist/promptrec/projects/demo-blommor/text.txt
	echo "02	ros	 Säg en blomma till" >> dist/promptrec/projects/demo-blommor/text.txt
	echo "03	tulpan	 Säg en tredje blomma" >> dist/promptrec/projects/demo-blommor/text.txt
	echo "04	nejlika	 Säg en blomma en fjärde och sista gång" >> dist/promptrec/projects/demo-blommor/text.txt
	cp -r promptrec/static dist/promptrec

webrtc: init
	cd webrtc_demo; \
	GOOS=linux GOARCH=amd64 go build -o ../dist/webrtc_demo/webrtc_demo *go; \
	GOOS=windows GOARCH=amd64 go build -o ../dist/webrtc_demo/webrtc_demo.exe *go; \
	cp -r index.html README.md ../dist/webrtc_demo


zip: init audio_streaming promptrec webrtc
	mkdir -p dist
	cp README.md dist
	cp technical_report_may_2020.tex dist
	cd dist; pdflatex technical_report_may_2020.tex; pdflatex technical_report_may_2020.tex
	cd dist; zip -q -r audio_demo.zip audio_streaming promptrec webrtc_demo technical_report_may_2020.pdf README.md; \
	rm -fr audio_streaming webrtc_demo promptrec README.md technical_report.*
	@echo "Output build saved in dist/audio_demo.zip"

clean:
	rm -rf dist
