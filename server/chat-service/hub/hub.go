package hub

import (
	"context"
	"encoding/json"
	"log"

	"talking/server/chat-service/service"
)

type WSMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type SendMessagePayload struct {
	RoomID  string `json:"roomId"`
	Content string `json:"content"`
}

type NewMessagePayload struct {
	MessageID      string `json:"messageId"`
	RoomID         string `json:"roomId"`
	SenderID       string `json:"senderId"`
	SenderNickname string `json:"senderNickname"`
	Content        string `json:"content"`
	CreatedAt      string `json:"createdAt"`
}

type StatusChangePayload struct {
	UserID string `json:"userId"`
	Online bool   `json:"online"`
}

type Hub struct {
	// 온라인 상태인 유저들의 클라이언트 관리 (유저 ID당 복수 연결 지원)
	clients    map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	authClient *service.AuthServiceClient
}

func NewHub(authClient *service.AuthServiceClient) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		authClient: authClient,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
				// 유저가 최초로 온라인 상태가 되었을 때 브로드캐스트
				go h.BroadcastStatusChange(client.userID, true)
			}
			h.clients[client.userID][client] = true
			log.Printf("User %s connected. Client connection registered.", client.userID)

		case client := <-h.unregister:
			if connections, ok := h.clients[client.userID]; ok {
				if _, exists := connections[client]; exists {
					delete(connections, client)
					close(client.send)
					log.Printf("Client connection unregistered for user %s", client.userID)

					if len(connections) == 0 {
						delete(h.clients, client.userID)
						// 유저가 완전히 오프라인 상태가 되었을 때 브로드캐스트
						go h.BroadcastStatusChange(client.userID, false)
					}
				}
			}
		}
	}
}

func (h *Hub) BroadcastToRoom(ctx context.Context, roomID string, eventName string, payload interface{}) {
	// 1. gRPC로 해당 룸 멤버 목록 조회
	memberIDs, err := h.authClient.GetRoomMembers(ctx, roomID)
	if err != nil {
		log.Printf("gRPC GetRoomMembers error for room %s: %v", roomID, err)
		return
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return
	}

	wsMsg := WSMessage{
		Event: eventName,
		Data:  payloadBytes,
	}

	msgBytes, err := json.Marshal(wsMsg)
	if err != nil {
		return
	}

	// 2. 룸 내에 접속해 있는 멤버들의 모든 소켓 커넥션에 메시지 푸시
	for _, memberID := range memberIDs {
		if connections, ok := h.clients[memberID]; ok {
			for client := range connections {
				select {
				case client.send <- msgBytes:
				default:
					close(client.send)
					delete(connections, client)
					if len(connections) == 0 {
						delete(h.clients, memberID)
						go h.BroadcastStatusChange(memberID, false)
					}
				}
			}
		}
	}
}

func (h *Hub) BroadcastStatusChange(userID string, online bool) {
	payload := StatusChangePayload{
		UserID: userID,
		Online: online,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}
	wsMsg := WSMessage{
		Event: "status_change",
		Data:  payloadBytes,
	}
	msgBytes, err := json.Marshal(wsMsg)
	if err != nil {
		return
	}

	// 모든 온라인 사용자에게 해당 상태 전송 (프론트가 알아서 친구만 필터링)
	for _, connections := range h.clients {
		for client := range connections {
			if client.userID == userID {
				continue
			}
			select {
			case client.send <- msgBytes:
			default:
				// Non-blocking skip
			}
		}
	}
}
