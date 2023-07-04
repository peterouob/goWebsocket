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
