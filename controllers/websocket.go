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

	websocket.Mutex.Lock()
	websocket.Clients[userID] = conn
	websocket.Mutex.Unlock()

	defer func() {
		websocket.Mutex.Lock()
		delete(websocket.Clients, userID)
		websocket.Mutex.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()

		if err != nil {
			break
		}
	}
}
