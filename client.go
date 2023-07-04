package main

import (
	"github.com/gorilla/websocket"
	"log"
)

// ClientList 客戶端列表，上線狀態
type ClientList map[*Client]bool

// Client 處理所有單用戶相關的內容
type Client struct {
	connection *websocket.Conn
	manager    *Manager

	//egress 避免客戶端併發權限，使用一個無緩衝的通道來防止連接同時獲得過多的請求
	egress chan []byte
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		//之所以會使用manager是因為會將一些事情引導到manager進行處理，例如像其他用戶廣播
		manager: manager,
		egress:  make(chan []byte),
	}
}

func (c *Client) readMessages() {
	//這邊使用defer是因為跳出for迴圈後再執行
	defer func() {
		//cleanup connection from client List，幫助我們清理未使用的客戶端
		c.manager.removeClient(c)
	}()

	for {
		//payload為負載，類行為byte
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
		log.Println(string(payload))
	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	for {

	}
}
