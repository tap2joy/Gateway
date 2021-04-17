package server

import (
	"context"

	"github.com/tap2joy/Gateway/services"
	pb "github.com/tap2joy/Protocols/go/grpc/gateway"
)

type Server struct {
}

// 推送消息
func (*Server) PushMessage(ctx context.Context, req *pb.PushMessageRequest) (*pb.PushMessageResponse, error) {
	senderName := req.SenderName
	content := req.Content
	userNames := req.UserNames
	timestamp := req.Timestamp

	err := services.GetChatMgr().PushMessage(senderName, content, timestamp, userNames)
	if err != nil {
		return nil, err
	}

	resp := new(pb.PushMessageResponse)
	return resp, nil
}

// 踢人
func (*Server) KickUser(ctx context.Context, req *pb.KickUserRequest) (*pb.KickUserResponse, error) {
	name := req.Name
	gate := req.Gate

	err := services.GetChatMgr().KickOutUser(name, gate)
	if err != nil {
		return nil, err
	}

	resp := new(pb.KickUserResponse)
	return resp, nil
}
