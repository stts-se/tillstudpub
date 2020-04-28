Simple server/client library for testing audio streaming using the MediaRecorder API.

To start the server, run

 `go run . `

Clients:

* Javascript: Point your browser to http://localhost:7651

* `Go` command line client: See folder `gocli`

Recorded audio is saved in the `data` folder. The last recorded file is always saved as `data/latest.raw`. To play a recorded `.raw` files, run play with the correct parameters, e.g.

 `play -e signed-integer -r 44100 -b 16 <rawfile>`

(See playraw_example.sh.)
