package login

import (
	"time"
)

// Users is a hardwired list of available users
var Users = []string{
	"gäst",
}

// Projects is a hardwired list of available projects
var Projects = []string{
	"tillstud_demo",
}

var timestamp = time.Now().Format("2006-01-02 15:04")

// Sessions is a hardwired list of available sessions
var Sessions = []string{
	"demo_session",
	timestamp,
}
