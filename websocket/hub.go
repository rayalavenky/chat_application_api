package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

var Clients = make(map[string]*websocket.Conn)

var Mutex sync.RWMutex

func AddClient(userID string, conn *websocket.Conn) {
	Mutex.Lock()
	defer Mutex.Unlock()

	Clients[userID] = conn
}

func RemoveClient(userID string) {
	Mutex.Lock()
	defer Mutex.Unlock()

	delete(Clients, userID)
}

func SendToUser(userID string, payload interface{}) {
	Mutex.RLock()
	conn, exists := Clients[userID]
	Mutex.RUnlock()

	if !exists {
		return
	}

	err := conn.WriteJSON(payload)
	if err != nil {
		conn.Close()

		Mutex.Lock()
		delete(Clients, userID)
		Mutex.Unlock()
	}
}
