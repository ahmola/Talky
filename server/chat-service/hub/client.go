package hub

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 개발 편의성을 위해 모든 Origin 허용
	},
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   string
	username string
	nickname string
}

func NewClient(hub *Hub, conn *websocket.Conn, userID, username, nickname string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		username: username,
		nickname: nickname,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket connection closed abruptly: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("WS message json unmarshal error: %v", err)
			continue
		}

		switch wsMsg.Event {
		case "send_message":
			var payload SendMessagePayload
			if err := json.Unmarshal(wsMsg.Data, &payload); err != nil {
				log.Printf("SendMessage payload unmarshal error: %v", err)
				continue
			}

			// 1. gRPC를 통해 auth-service에 메시지 저장 요청
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			res, err := c.hub.authClient.SaveMessage(ctx, payload.RoomID, c.userID, payload.Content)
			cancel()

			if err != nil {
				log.Printf("SaveMessage gRPC call failed: %v", err)
				c.sendError("SAVE_FAILED", "메시지를 전송하지 못했습니다.")
				continue
			}

			// 2. 방 참여자들에게 실시간 브로드캐스트
			ctx = context.Background()
			c.hub.BroadcastToRoom(ctx, payload.RoomID, "new_message", NewMessagePayload{
				MessageID:      res.MessageId,
				RoomID:         payload.RoomID,
				SenderID:       c.userID,
				SenderNickname: res.SenderNickname,
				Content:        payload.Content,
				CreatedAt:      res.CreatedAt,
			})
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 큐에 쌓여 있는 추가 메시지들도 결합하여 전송
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendError(code, message string) {
	errPayload := map[string]string{
		"code":    code,
		"message": message,
	}
	payloadBytes, _ := json.Marshal(errPayload)
	wsMsg := WSMessage{
		Event: "error",
		Data:  payloadBytes,
	}
	msgBytes, _ := json.Marshal(wsMsg)
	c.send <- msgBytes
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, userID, username, nickname string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade failed: %v", err)
		return
	}

	client := NewClient(hub, conn, userID, username, nickname)
	hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}
