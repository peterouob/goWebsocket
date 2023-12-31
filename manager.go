// http status 101 代表正在切換協議
package main

// manager.go // 管理websocket
import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

// 管理websocket
var (
	websocketUpgrade = websocket.Upgrader{
		//cross
		CheckOrigin: checkOrigin,
		//確保用戶不會發送巨大封包
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Manager struct {
	clients ClientList
	//會有很多人同時連接到API，使用互斥鎖保護
	sync.RWMutex

	otps RetentionMap

	//將type當作key並允許我們獲取事件處理程序
	handlers map[string]EventHandler
}

// 工廠模式
func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:  make(ClientList),
		otps:     NewRetentionMap(ctx, 5*time.Second),
		handlers: make(map[string]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
	m.handlers[EventChangeRoom] = ChangeRoomHandler
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	//檢查事件類型是否是處理程序的一部分，處理程序是一個使用事件類型作為key的map
	//因此每當我們收到類型設置為發送消息，都會觸發發送消息
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("there is no such event type")
	}
}

func (m *Manager) serveWs(w http.ResponseWriter, r *http.Request) {
	//驗證OTP是否有效
	otp := r.URL.Query().Get("otp")
	if otp == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !m.otps.VerifyOTP(otp) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
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

func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request) {
	type userLoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var req userLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//form
	if req.Username == "test" && req.Password == "test" {
		type Response struct {
			OTP string `json:"otp"`
		}
		otp := m.otps.NewOTP()
		resp := Response{
			OTP: otp.Key,
		}
		data, err := json.Marshal(resp)
		if err != nil {
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
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

// Cross Origin
func checkOrigin(r *http.Request) bool {
	//true將會連接，false關閉連接
	origin := r.Header.Get("Origin")
	switch origin {
	case "http://localhost:8082":
		return true
	default:
		return false
	}
}
