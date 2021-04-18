package services

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
	"unsafe"

	"github.com/tap2joy/Gateway/utils"
	pb_common "github.com/tap2joy/Protocols/go/common"
	pb "github.com/tap2joy/Protocols/go/gateway"
	pb_chat "github.com/tap2joy/Protocols/go/grpc/chat"
	pb_gate "github.com/tap2joy/Protocols/go/grpc/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var instance *ChatMgr
var once sync.Once

// GetServiceMgr 获取单例
func GetChatMgr() *ChatMgr {
	once.Do(func() {
		if instance == nil {
			instance = NewChatMgr()
		}
	})
	return instance
}

type ChatMgr struct {
	ChatServers  []string                // 聊天服务器列表
	UsersConn    map[string]*net.TCPConn // 玩家名字与链接映射
	Conn2User    map[string]string       // 链接到用户映射
	UserChannels map[string]uint32       // 聊天频道映射
}

func NewChatMgr() *ChatMgr {
	mgr := &ChatMgr{
		UsersConn:    make(map[string]*net.TCPConn),
		Conn2User:    make(map[string]string),
		UserChannels: make(map[string]uint32),
	}

	mgr.RefreshChatService()
	return mgr
}

// 刷新聊天服务列表
func (mgr *ChatMgr) RefreshChatService() {
	resp, err := GetChatServices()
	if err != nil {
		fmt.Printf("get chat service failed, err =%v\n", err)
	} else {
		fmt.Println("get chat service success")

		mgr.ChatServers = make([]string, 0)
		for _, v := range resp.List {
			mgr.ChatServers = append(mgr.ChatServers, v)
		}
	}

	fmt.Printf("%v\n", mgr.ChatServers)
}

func (mgr *ChatMgr) Init() {
	fmt.Printf("chat %s init\n", utils.GetLocalAddress())
}

// 推送消息
func (mgr *ChatMgr) PushMessage(senderName string, content string, timestamp uint64, userNames []string) error {
	msg := &pb.SPushMessage{
		SenderName: senderName,
		Content:    content,
		Timestamp:  timestamp,
	}

	msgByte, err := proto.Marshal(msg)
	if err != nil {
		fmt.Printf("Push message Marshal error %s\n", err.Error())
		return nil
	}

	buf := &bytes.Buffer{}
	var head []byte
	head = make([]byte, 8)
	dataLen := uint32(len(msgByte)) //uint32(bytes.Count(msgByte, nil)-1)
	binary.BigEndian.PutUint32(head[0:4], dataLen)
	binary.BigEndian.PutUint32(head[4:8], uint32(pb_common.Mid_G2C_PUSH_MESSAGE))
	buf.Write(head[:8])
	buf.Write(msgByte)

	for _, v := range userNames {
		if conn, ok := mgr.UsersConn[v]; ok {
			(*conn).Write(buf.Bytes())
		}
	}

	return nil
}

// 发送聊天消息
func (mgr *ChatMgr) SendMessage(channelId uint32, senderName string, content string, isSystem bool) (string, error) {
	result := ""
	if !isSystem {
		_, ok := mgr.UsersConn[senderName]
		if !ok {
			fmt.Printf("user %s conn not exist\n", senderName)
			return result, status.Errorf(codes.Internal, "conn not exist")
		}
	}

	chatServicesCount := len(mgr.ChatServers)
	if chatServicesCount == 0 {
		return result, status.Errorf(codes.Internal, "no chat service found")
	}

	// 随机一个chatService
	randIndex := rand.Intn(chatServicesCount)
	chatAddress := mgr.ChatServers[randIndex]

	conn, err := grpc.Dial(chatAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return result, err
	}
	defer conn.Close()
	c := pb_chat.NewChatServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := c.SendMessage(ctx, &pb_chat.SendMessageRequest{Channel: channelId, SenderName: senderName, Content: content, System: isSystem})
	if err != nil {
		return result, err
	}

	return resp.Result, nil
}

// 用户登陆
func (mgr *ChatMgr) UserLogin(name string, channelId uint32, conn *net.Conn) error {
	localAddress := utils.GetLocalAddress()

	resp, err := UserOnline(name, localAddress, channelId)
	if err != nil {
		fmt.Printf("user %s login failed\n", name)
		return err
	}

	if resp.OldUser != nil {
		// todo: kick old user
		if resp.OldUser.Gateway == localAddress {
			mgr.KickOutUser(resp.OldUser.Name, localAddress)
		}
	}

	mgr.UserChannels[name] = channelId

	tcpConn := (*conn).(*net.TCPConn)
	mgr.UsersConn[name] = tcpConn
	mgr.Conn2User[ConnPointer2String(tcpConn)] = name
	fmt.Printf("user %s login success\n", name)
	return nil
}

// 用户下线
func (mgr *ChatMgr) UserLogout(name string) error {
	if _, ok := mgr.UsersConn[name]; ok {
		err := UserOffline(name)
		if err != nil {
			return err
		}

		connStr := ConnPointer2String(mgr.UsersConn[name])
		delete(mgr.Conn2User, connStr)
		(*mgr.UsersConn[name]).Close()
		delete(mgr.UsersConn, name)
		fmt.Printf("user %s logout success\n", name)

		notifyMsg := fmt.Sprintf("user [%s] leave ...", name)
		mgr.SendMessage(mgr.UserChannels[name], "system", notifyMsg, true)
		delete(mgr.UserChannels, name)
	}

	return nil
}

// 链接断开
func (mgr *ChatMgr) OnConnClosed(conn *net.TCPConn) {
	connStr := ConnPointer2String(conn)
	if name, ok := mgr.Conn2User[connStr]; ok {
		fmt.Printf("user %s conn closed\n", name)
		mgr.UserLogout(name)
	}
}

// 踢出用户
func (mgr *ChatMgr) KickOutUser(name string, gateAddress string) error {
	localAddress := utils.GetLocalAddress()
	if gateAddress == localAddress {
		if _, ok := mgr.UsersConn[name]; ok {
			// 发一条顶号通知
			userNames := []string{name}
			mgr.PushMessage("system", "you are login in another place ...", uint64(time.Now().Unix()), userNames)

			connStr := ConnPointer2String(mgr.UsersConn[name])
			delete(mgr.Conn2User, connStr)
			(*mgr.UsersConn[name]).Close()
			delete(mgr.UsersConn, name)
			fmt.Printf("kick out user %s success\n", name)
		}
	} else {
		conn, err := grpc.Dial(gateAddress, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return err
		}
		defer conn.Close()
		c := pb_gate.NewGatewayServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err = c.KickUser(ctx, &pb_gate.KickUserRequest{Name: name, Gate: gateAddress})
		if err != nil {
			return err
		}
	}

	return nil
}

// 获取聊天记录
func (mgr *ChatMgr) GetChatLogs(channelId uint32) (*pb_chat.GetChatLogResponse, error) {
	chatServicesCount := len(mgr.ChatServers)
	if chatServicesCount == 0 {
		return nil, status.Errorf(codes.Code(pb_common.ErrorCode_SERVICE_NOT_EXIST_ERROR), "service not exist")
	}

	// 随机一个chatService
	randIndex := rand.Intn(chatServicesCount)
	chatAddress := mgr.ChatServers[randIndex]

	conn, err := grpc.Dial(chatAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := pb_chat.NewChatServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := c.GetChatLog(ctx, &pb_chat.GetChatLogRequest{Channel: channelId})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// 获取频道列表
func (mgr *ChatMgr) GetChannelList() (*pb_chat.GetChannelListResponse, error) {
	chatServicesCount := len(mgr.ChatServers)
	if chatServicesCount == 0 {
		return nil, status.Errorf(codes.Code(pb_common.ErrorCode_SERVICE_NOT_EXIST_ERROR), "service not exist")
	}

	// 随机一个chatService
	randIndex := rand.Intn(chatServicesCount)
	chatAddress := mgr.ChatServers[randIndex]

	conn, err := grpc.Dial(chatAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := pb_chat.NewChatServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := c.GetChannelList(ctx, &pb_chat.GetChannelListRequest{})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (mgr *ChatMgr) ChangeChannel(name string, channelId uint32) {
	mgr.UserChannels[name] = channelId
}

func ConnPointer2String(conn *net.TCPConn) string {
	strPointerHex := fmt.Sprintf("%p", unsafe.Pointer(conn))
	return strPointerHex
}
