# promptrec

A simple demo for recording text prompts on a project basis.

Start server:
$ go run .

Point your browser to http://localhost:3092, choose a project, enter session and user name, and start recording.

Output files are saved in the `projects` folder:

    projectname/
      sessionname/
        username/
          audiofiles and accompanying jsonfiles

If you prefer precompiled executables command from a [published release](https://github.com/stts-se/tillstudpub/releases):

    $ unzip audio_demo.zip
    $ cd promptrec
    $ ./promptrec_server


## Define a new project

Project data is stored in the `projects` folder. To create a new project, create a folder inside the `projects` folder. The folder name will be the name of the project. Inside the project folder, create a text file `text.txt` with the prompt texts, one prompt per line on the following format:

1. prompt id
2. prompt text (to be read)
3. instructions to the user (about the prompt or how it should be read) (optinal)

## Dependencies

ffmpeg
