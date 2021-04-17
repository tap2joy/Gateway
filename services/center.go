package services

import (
	"context"
	"time"

	"github.com/tap2joy/Gateway/utils"
	protocols "github.com/tap2joy/Protocols/go/grpc/center"
	"google.golang.org/grpc"
)

// 注册gateway到centerService
func RegisterGatewayService(gateAddress string) error {
	address := utils.GetString("grpc", "grpc_center_address")

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}
	defer conn.Close()
	c := protocols.NewCenterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.RegisterService(ctx, &protocols.RegisterServiceRequest{Type: "gateway", Address: gateAddress})
	if err != nil {
		return err
	}

	return nil
}

// 获取聊天服务列表
func GetChatServices() (*protocols.GetServicesResponse, error) {
	address := utils.GetString("grpc", "grpc_center_address")

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := protocols.NewCenterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := c.GetServices(ctx, &protocols.GetServicesRequest{Type: "chat"})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// 用户上线
func UserOnline(name string, gateAddress string, channelId uint32) (*protocols.UserOnlineResponse, error) {
	address := utils.GetString("grpc", "grpc_center_address")

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := protocols.NewCenterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := c.UserOnline(ctx, &protocols.UserOnlineRequest{Name: name, Gate: gateAddress, Channel: channelId})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// 用户下线
func UserOffline(name string) error {
	address := utils.GetString("grpc", "grpc_center_address")

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}
	defer conn.Close()
	c := protocols.NewCenterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.UserOffline(ctx, &protocols.UserOfflineRequest{Name: name})
	if err != nil {
		return err
	}

	return nil
}

func ChangeChannel(name string, channelId uint32) error {
	address := utils.GetString("grpc", "grpc_center_address")

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}
	defer conn.Close()
	c := protocols.NewCenterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.ChangeChannel(ctx, &protocols.ChangeChannelRequest{Name: name, Channel: channelId})
	if err != nil {
		return err
	}

	return nil
}
