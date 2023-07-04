// http status 101 代表正在切換協議
package main

// manager.go // 管理websocket
import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// 管理websocket
var (
	websocketUpgrade = websocket.Upgrader{
		//確保用戶不會發送巨大封包
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Manager struct {
}

// 工廠模式
func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) serveWs(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection")
	//upgrade regular http connection into websocket
	conn, err := websocketUpgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	conn.Close()
}
