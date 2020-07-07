# rispik

This folder contains code and documentation for the development of a first recording tool, to be used in test recordings sessions in August, 2020.


# Description

## User Web Clients

    - HTML5 + CSS Grid, following separate layout spec.
    - Simple login: USER + SESSION
    - Session timer
    - Start/stop rec buttons
    - Audio input meter
    - Server ASR result text area, that scrolls down, pushing earlier results at bottom of list 
    - Editable ASR result text area
    - Client starts audio recording
    - Client connects to server over WebSocket
    - Client-Server handshake (wearing gloves, of course)
    - Client streams audio to server over WebSocket
    - Client listens for ASR response on WebSocket
    

## CLI client
    - Stream audio (live from mic using "rec" (SoX) or from audio file) from command line to server via WebSocket
    - Read ASR result over WebSocket

## Admin client

    - Simple log page, to which server sends log messages over WebSocket
    - File listings?
    - Client listings? (See log above)

## Server

    - Serves client and admin GUI
    - Waits for WebSocket calls
    - Reads streaming audio over WebSocket
    - Saves streaming audio locally (or on remote location?)
    - Saves meta data along with audio
    - Optionally streams incoming audio stream to stand-alone ASR server
    - If streaming to ASR server, sends corresponding ASR result to user client over WebSocket
    - Keeps track of clients

## Simple flow chart

![Simple flow chart](rispik_chart.png)


