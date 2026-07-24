package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"talking/server/auth-service/handler"
	"talking/server/auth-service/repository"
	"talking/server/auth-service/service"
	pb "talking/server/proto/auth"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/talking?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	schemaPath := os.Getenv("SCHEMA_PATH")
	if schemaPath == "" {
		schemaPath = "./sql/postgres/schema.sql"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "talking-super-secret-key"
	}

	// 1. 데이터베이스 커넥션 생성
	db, err := repository.NewDatabase(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Printf("Initializing database schema from %s...", schemaPath)
	if err := db.InitSchema(schemaPath); err != nil {
		log.Printf("Warning: database schema initialization had issues: %v", err)
	}

	// 2. 인증 모듈 초기화
	authProvider := service.NewJWTAuthProvider(jwtSecret, 24*time.Hour)

	// 3. HTTP REST API 라우터 구축
	h := handler.NewHandler(db, authProvider)
	r := gin.Default()

	// CORS 미들웨어 구성
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		api.POST("/auth/register", h.Register)
		api.POST("/auth/login", h.Login)

		// 인증 보호 그룹
		authorized := api.Group("")
		authorized.Use(h.AuthMiddleware())
		{
			authorized.GET("/friends", h.GetFriends)
			authorized.POST("/friends", h.AddFriend)
			authorized.DELETE("/friends/:userId", h.DeleteFriend)

			authorized.GET("/rooms", h.GetRooms)
			authorized.POST("/rooms", h.CreateRoom)
			authorized.GET("/rooms/:roomId/messages", h.GetRoomMessages)
		}
	}

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Starting HTTP REST Server on port %s...", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// 4. gRPC 서버 구축
	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on TCP for gRPC: %v", err)
	}

	grpcServer := grpc.NewServer()
	authGrpcService := service.NewGRPCServer(db, authProvider)
	pb.RegisterAuthServiceServer(grpcServer, authGrpcService)

	go func() {
		log.Printf("Starting gRPC Server on port %s...", grpcPort)
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Graceful Shutdown 리스너
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP Server forced to shutdown: %v", err)
	}

	grpcServer.GracefulStop()
	log.Println("Servers exited properly.")
}
