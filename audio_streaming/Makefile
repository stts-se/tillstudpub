.PHONY: all
all: audio_streaming audio_streaming_mac audio_streaming_win.exe audio_streaming.zip

audio_streaming: main.go static/*css static/*html static/*js static/*ico static/*png README.md
	GOOS=linux GOARCH=amd64 go build -o audio_streaming main.go

audio_streaming_mac: main.go static/*css static/*html static/*js static/*ico static/*png README.md
	GOOS=darwin GOARCH=amd64 go build -o audio_streaming_mac main.go


audio_streaming_win.exe: main.go static/*css static/*html static/*js static/*ico static/*png README.md
	GOOS=windows GOARCH=amd64 go build -o audio_streaming_win.exe main.go


audio_streaming.zip: audio_streaming audio_streaming_mac audio_streaming_win.exe static/*css static/*html static/*js static/*ico static/*png README.md
	zip -q audio_streaming.zip audio_streaming audio_streaming_mac audio_streaming_win.exe static/*css static/*html static/*js static/*ico static/*png README.md
	rm audio_streaming
	rm audio_streaming_mac
	rm audio_streaming_win.exe


clean:
	rm -f audio_streaming audio_streaming_mac audio_streaming_win.exe audio_streaming.zip
