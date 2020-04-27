Simple server/client library for testing audio streaming using the MediaRecorder API.

To start the server, run

 `go run . `

To play .raw files, run play with the correct parameters, e.g.

 `/usr/bin/play -e signed-integer -r 44100 -b 16 <filename>`


Clients:

	Javascript: Point your browser to http://localhost:7654

	`Go` client: See folder `gocli`
