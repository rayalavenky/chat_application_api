package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

var Clients = make(map[string]*websocket.Conn)

var Mutex sync.Mutex
