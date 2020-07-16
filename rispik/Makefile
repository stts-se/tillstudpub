.PHONY: all
all: rispik zip

rispik: 
	GOOS=linux GOARCH=amd64 go build -o rispik *go
	GOOS=windows GOARCH=amd64 go build -o rispik.exe *go

zip: rispik
	zip -q -r rispik.zip rispik.exe rispik static/* README.md docs/*; \
	rm -fr rispik.exe rispik

clean:
	rm -rf dist
