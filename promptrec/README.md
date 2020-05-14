# promptrec

A simple demo for recording text prompts on a project basis.

Start server:
$ go run .

Point your browser to http://localhost:3092, choose a project, enter session and user name, and start recording.

Output files are saved in the `projects` folder (one folder per project).

If you prefer precompiled executables command from a [published release](https://github.com/stts-se/tillstudpub/releases):

    $ unzip audio_demo.zip
    $ cd promptrec
    $ ./promptrec_server


The precompiled release includes a demo project for testing, `demo-blommor`. If you start the server outside of a release, your initial project list may be empty, and you need to create a new project.

## Data structure

Project data is stored in the `projects` folder, using the following file structure:

    projects/
       <projectname>/
         text.txt
         <sessionname>/
           <username>/
             user's audio files and accompanying json files with metadata


## Define a new project

To create a new project, create a folder inside the `projects` folder. The folder name will be the name of the project. Inside the project folder, create a text file `text.txt` with the prompt texts, one prompt per line on the following format (tab-separated):

1. prompt id
2. prompt text (to be read)
3. instructions, if any (directed to the user, about the prompt or how it should be read)

## Dependencies

ffmpeg
