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
### 使用Event判斷消息類型並取代原先的發送和接受
```javascript
        <script>
            let selectedChat = "general";
            class Event {
                //更好的控制用戶發送的訊息
                constructor(type ,payload) {
                    this.type = type;
                    this.payload = payload;
                }
            }
            function routeEvent(event){
                if(event.type === undefined){
                    alert("no type field in the event");
                }
                switch (event.type){
                    case "new_message":
                        console.log("new Message");
                        break;
                    default:
                        alert("not supported this type");
                        break;
                }
            }

            function sendEvent(eventName,payload){
                const event = new Event(eventName,payload);
                conn.send(JSON.stringify(event))
            }

            function changeChatRoom (){
                let newChat = document.getElementById("chatroom");
                if(newChat !== null && newChat.value !== selectedChat){
                    console.log(newChat);
                }
                return false;
            }
            function sendMessage(){
                let newMessage = document.getElementById("message");
                if(newMessage !== null ){
                    // console.log(newMessage);
                    // conn.send(newMessage.value);
                    sendEvent("send_message", newMessage)
                }
                return false;
            }

            window.onload = function(){
                document.getElementById("chatroom-selection").onsubmit = changeChatRoom;
                document.getElementById("chatroom-message").onsubmit = sendMessage;

                if(window["WebSocket"]){
                    console.log("support websocket")
                    //connect to websocket
                    conn = new WebSocket("ws://"+document.location.host+"/ws")
                    conn.onmessage = function(e){
                        // console.log(e)
                        const eventData = JSON.parse(e.data);
                        const event = Object.assign(new Event, eventData);
                        routeEvent(event)
                    }
                }else{
                    alert("not support websocket")
                }
            }
        </script>
```
```go
// event.go
package main

import "encoding/json"

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	//使用RawMessage（原始格式）是因為我們希望使用者可以發送任何類型
}

type EventHandler func(event Event, c *Client) error

const (
	EventSendMessage = "send_message"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

//將這些儲存在管理器，確保管理器方便處理
```

```go
type Manager struct {
	clients ClientList
	//會有很多人同時連接到API，使用互斥鎖保護
	sync.RWMutex
	//將type當作key並允許我們獲取事件處理程序
	handlers map[string]EventHandler
}

// 工廠模式
func NewManager() *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
}

func SendMessage(event Event, c *Client) error {
	fmt.Println(event)
	return nil
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
```

```go
func (c *Client) readMessages() {
	//這邊使用defer是因為跳出for迴圈後再執行
	defer func() {
		//cleanup connection from client List，幫助我們清理未使用的客戶端
		c.manager.removeClient(c)
	}()

	for {
		//payload為負載，類行為byte
		_, payload, err := c.connection.ReadMessage()
		//messageType在RFC中定義有幾種不同的消息類型讓你對數據,二進制進行ping/pong
		if err != nil {
			//連接意外關閉返回錯誤
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				//檢查異常是因為不希望正常斷開的時候也被當成錯誤紀錄
				log.Printf("error reading message %v", err)
				break
			}
		}

		//測試寫入egress，每read一個message都會發送給其他所有客戶
		//for wsclient := range c.manager.clients {
		//	wsclient.egress <- payload
		//}
		//
		//log.Println(messageType)
		//log.Println(string(payload)) ->使用Event取代
		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error unmarshalling :%v", err)
			break
		}
		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("error handling message", err)
		}
	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed :", err)
				}
				return //return 後會觸發破壞迴圈觸發defer
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				break
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("failed to send message %v:", err)
			}
			log.Println("message sent")
		}
	}
}
```