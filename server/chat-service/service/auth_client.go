package service

import (
	"context"

	pb "talking/server/proto/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthServiceClient struct {
	client pb.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthServiceClient(addr string) (*AuthServiceClient, error) {
	// 컨테이너/내부 망 통신이므로 insecure 크레덴셜 사용
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &AuthServiceClient{
		client: pb.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *AuthServiceClient) Close() error {
	return c.conn.Close()
}

func (c *AuthServiceClient) VerifyToken(ctx context.Context, token string) (*pb.VerifyTokenResponse, error) {
	return c.client.VerifyToken(ctx, &pb.VerifyTokenRequest{Token: token})
}

func (c *AuthServiceClient) GetRoomMembers(ctx context.Context, roomID string) ([]string, error) {
	res, err := c.client.GetRoomMembers(ctx, &pb.GetRoomMembersRequest{RoomId: roomID})
	if err != nil {
		return nil, err
	}
	return res.UserIds, nil
}

func (c *AuthServiceClient) SaveMessage(ctx context.Context, roomID, senderID, content string) (*pb.SaveMessageResponse, error) {
	return c.client.SaveMessage(ctx, &pb.SaveMessageRequest{
		RoomId:   roomID,
		SenderId: senderID,
		Content:  content,
	})
}
