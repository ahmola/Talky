package service

import (
	"context"
	"time"

	"talking/server/auth-service/repository"
	pb "talking/server/proto/auth"
)

type GRPCServer struct {
	pb.UnimplementedAuthServiceServer
	db           *repository.Database
	authProvider AuthProvider
}

func NewGRPCServer(db *repository.Database, authProvider AuthProvider) *GRPCServer {
	return &GRPCServer{
		db:           db,
		authProvider: authProvider,
	}
}

func (s *GRPCServer) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	payload, err := s.authProvider.VerifyToken(req.Token)
	if err != nil {
		return nil, err
	}

	return &pb.VerifyTokenResponse{
		UserId:   payload.UserID,
		Username: payload.Username,
		Nickname: payload.Nickname,
	}, nil
}

func (s *GRPCServer) GetRoomMembers(ctx context.Context, req *pb.GetRoomMembersRequest) (*pb.GetRoomMembersResponse, error) {
	members, err := s.db.GetRoomMembers(ctx, req.RoomId)
	if err != nil {
		return nil, err
	}

	return &pb.GetRoomMembersResponse{
		UserIds: members,
	}, nil
}

func (s *GRPCServer) SaveMessage(ctx context.Context, req *pb.SaveMessageRequest) (*pb.SaveMessageResponse, error) {
	msg, err := s.db.SaveMessage(ctx, req.RoomId, req.SenderId, req.Content)
	if err != nil {
		return nil, err
	}

	return &pb.SaveMessageResponse{
		MessageId:      msg.ID,
		CreatedAt:      msg.CreatedAt.Format(time.RFC3339),
		SenderNickname: msg.SenderNickname,
	}, nil
}

func (s *GRPCServer) CheckFriendship(ctx context.Context, req *pb.CheckFriendshipRequest) (*pb.CheckFriendshipResponse, error) {
	isFriend, err := s.db.IsFriend(ctx, req.UserId, req.FriendId)
	if err != nil {
		return nil, err
	}

	return &pb.CheckFriendshipResponse{
		IsFriend: isFriend,
	}, nil
}
