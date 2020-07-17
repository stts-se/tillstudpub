package main

import (
	"fmt"
	"net/http"

	"github.com/stts-se/tillstudpub/rispik/logger"
)

func openAdminWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		msg := fmt.Sprintf("failed to upgrade to ws: %v", err)
		httpError(w, msg, "Failed to upgrade to ws", http.StatusInternalServerError)
		return
	}
	logger.AddWSListener(conn)
	logger.Info("Registered admin socket")
}
