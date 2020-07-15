package login

import (
	"time"
)

var Users = []string{
	"gäst",
}

var Projects = []string{
	"tillstud_demo",
}

var timestamp = time.Now().Format("2006-01-02 15:04")

var Sessions = []string{
	"demo_session",
	timestamp,
}
