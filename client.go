package main

import "github.com/gorilla/websocket"

// 客戶端列表
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
