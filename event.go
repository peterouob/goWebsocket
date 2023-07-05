package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	//使用RawMessage（原始格式）是因為我們希望使用者可以發送任何類型
}

type EventHandler func(event Event, c *Client) error

const (
	EventSendMessage = "send_message"
	EventNewMessage  = "new_message"
	EventChangeRoom  = "change_room"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}
type ChangeRoomEvent struct {
	Name string `json:"name"`
}

func ChangeRoomHandler(event Event, c *Client) error {
	var chatRoomEvent ChangeRoomEvent
	if err := json.Unmarshal(event.Payload, &chatRoomEvent); err != nil {
		return fmt.Errorf("bad payload in request : %v", err)
	}
	c.chatroom = chatRoomEvent.Name
	return nil
}
func SendMessage(event Event, c *Client) error {
	//fmt.Println(event)
	var chatevent SendMessageEvent
	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("bad payload in request : %v", err)
	}
	var broadMessage NewMessageEvent
	broadMessage.Sent = time.Now()
	broadMessage.Message = chatevent.Message
	broadMessage.From = chatevent.From

	data, err := json.Marshal(broadMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message : %v", err)
	}
	var outgoing Event
	outgoing.Type = EventNewMessage
	outgoing.Payload = data
	for clients := range c.manager.clients {
		if clients.chatroom == c.chatroom {
			clients.egress <- outgoing
		}
	}
	return nil
}

//將這些儲存在管理器，確保管理器方便處理
