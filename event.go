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
