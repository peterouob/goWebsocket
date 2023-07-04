## golangWebsocket

### 處理html靜態檔案和websocket路由

```go
package main

import (
	"log"
	"net/http"
)

func main() {
	setupApi()
	log.Fatalln(http.ListenAndServe(":8082", nil))
}

func setupApi() {
	manger := Manager{}
	http.Handle("/", http.FileServer(http.Dir("./front"))) //加載前端
	http.HandleFunc("/ws", manger.serveWs)
}

//go run *.go 運行所有文件
```
### 新增一個簡單的ws服務
```go
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
```

### 新增基本客戶端
```go
package main

import "github.com/gorilla/websocket"

//客戶端列表
type ClientList map[*Client]bool

//處理所有單用戶相關的內容
type Client struct {
	connection *websocket.Conn
	manager *Manager
}

func NewClient(conn *websocket.Conn,manager *Manager) *Client {
	return &Client{
		connection: conn,
		//之所以會使用manager是因為會將一些事情引導到manager進行處理，例如像其他用戶廣播
		manager: manager,
	}
}
```

### 修改Manager讓Manager能夠維護Client
```go
type Manager struct {
	clients ClientList
	//會有很多人同時連接到API，使用互斥鎖保護
	sync.RWMutex
}
```

### 新增和移除
```go
func (m *Manager) serveWs(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection")
	//upgrade regular http connection into websocket
	conn, err := websocketUpgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := NewClient(conn,m)
	m.addClient(client)
	//conn.Close()
}

//向管理起添加或刪除客戶端
func (m *Manager)addClient(client *Client) {
	m.Lock()
	//鎖上後當兩個人同時連接就不會同時修該Client list，Client list本質為map
	defer m.Unlock()
	m.clients[client] = true
}

func (m *Manager) removeClient(client *Client){
	m.Lock()
	defer m.Unlock()
	if _,ok := m.clients[client]; ok{
		client.connection.Close()
		delete(m.clients, client)
	}
}
```
### 讀取訊息
```go
package main

import (
	"github.com/gorilla/websocket"
	"log"
)

// 客戶端列表，上線狀態
type ClientList map[*Client]bool

// 處理所有單用戶相關的內容
type Client struct {
	connection *websocket.Conn
	manager    *Manager
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		//之所以會使用manager是因為會將一些事情引導到manager進行處理，例如像其他用戶廣播
		manager: manager,
	}
}

// 使用一個無緩衝的通道來防止連接同時獲得過多的請求
func (c *Client) readMessages() {
	//這邊使用defer是因為跳出for迴圈後再執行
	defer func() {
		//cleanup connection
		c.manager.removeClient(c)
	}()

	for {
		//payload為負載
		messageType, payload, err := c.connection.ReadMessage()
		//messageType在RFC中定義有幾種不同的消息類型讓你對數據,二進制進行ping/pong
		if err != nil {
			//連接意外關閉返回錯誤
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				//檢查異常是因為不希望正常斷開的時候也被當成錯誤紀錄
				log.Printf("error reading message %v", err)
				break
			}
		}
		log.Println(messageType)
		log.Println(payload)
	}
}
```
```javascript
function sendMessage(){
                var newMessage = document.getElementById("message");
                if(newMessage !== null ){
                    // console.log(newMessage);
                    conn.send(newMessage.value);
                }
                return false;
            }
```