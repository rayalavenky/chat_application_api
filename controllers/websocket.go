package controllers

import (
	"chat_application_api/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
)

var upgrader = gorilla.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebSocketHandler(c *gin.Context) {

	userID := c.Query("userId")

	conn, err := upgrader.Upgrade(
		c.Writer,
		c.Request,
		nil,
	)

	if err != nil {
		return
	}

	websocket.AddClient(userID, conn)

	defer func() {
		websocket.RemoveClient(userID)
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()

		if err != nil {
			break
		}
	}
}
