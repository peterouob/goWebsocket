// http status 101 代表正在切換協議
package main

// manager.go // 管理websocket
import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
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
	clients ClientList
	//會有很多人同時連接到API，使用互斥鎖保護
	sync.RWMutex
}

// 工廠模式
func NewManager() *Manager {
	return &Manager{
		clients: make(ClientList),
	}
}

func (m *Manager) serveWs(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection")
	//upgrade regular http connection into websocket
	conn, err := websocketUpgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := NewClient(conn, m)
	m.addClient(client)
	// Start client processing
	go client.readMessages()
	go client.writeMessages()
}

// 向管理起添加或刪除客戶端
func (m *Manager) addClient(client *Client) {
	m.Lock()
	//鎖上後當兩個人同時連接就不會同時修該Client list，Client list本質為map
	defer m.Unlock()
	m.clients[client] = true
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}
}
