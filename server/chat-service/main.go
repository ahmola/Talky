package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"talking/server/chat-service/hub"
	"talking/server/chat-service/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	authGrpcAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authGrpcAddr == "" {
		authGrpcAddr = "localhost:50051"
	}

	// 1. auth-service에 연동할 gRPC 클라이언트 초기화
	log.Printf("Connecting to auth-service gRPC at %s...", authGrpcAddr)
	authClient, err := service.NewAuthServiceClient(authGrpcAddr)
	if err != nil {
		log.Fatalf("Failed to initialize auth gRPC client: %v", err)
	}
	defer authClient.Close()

	// 2. WebSocket Hub 생성 및 실행
	h := hub.NewHub(authClient)
	go h.Run()

	// 3. HTTP 라우트 설정 (/ws 경로에서 핸들셰이크)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Token is required", http.StatusUnauthorized)
			return
		}

		// gRPC를 통한 JWT 토큰 검증 요청
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		res, err := authClient.VerifyToken(ctx, token)
		if err != nil {
			log.Printf("Token verification failed: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// WebSocket 연결 수립 및 등록
		hub.ServeWs(h, w, r, res.UserId, res.Username, res.Nickname)
	})

	server := &http.Server{
		Addr: ":" + port,
	}

	go func() {
		log.Printf("Starting WebSocket chat-service on port %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("WebSocket server run error: %v", err)
		}
	}()

	// Graceful shutdown 리스너
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down chat-service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("WebSocket server forced to shutdown: %v", err)
	}

	log.Println("chat-service exited properly.")
}
