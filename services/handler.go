package services

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	pb_common "github.com/tap2joy/Protocols/go/common"
	pb "github.com/tap2joy/Protocols/go/gateway"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// 处理登陆
func HandleUserLogin(conn net.Conn, mid pb_common.Mid, msg interface{}) {
	packet := msg.(*pb.CLogin)
	fmt.Println("message handler mid:", mid, " body:", packet)

	name := packet.Name
	channel := packet.Channel
	err := GetChatMgr().UserLogin(name, channel, &conn)
	if err != nil {
		fmt.Printf("user login error %s\n", err.Error())
		PushErrorMessage(conn, err)
		return
	}

	// 回复登陆成功消息
	respMsg := &pb.SLogin{
		Name:    name,
		Channel: channel,
	}

	msgByte, err := proto.Marshal(respMsg)
	if err != nil {
		fmt.Printf("msg Marshal error %s\n", err.Error())
		return
	}

	SendPacket(conn, pb_common.Mid_G2C_USER_LOGIN, msgByte)

	notifyMsg := fmt.Sprintf("user [%s] enter room ...", name)
	GetChatMgr().SendMessage(packet.Channel, "system", notifyMsg, true)
}

// 处理用户登出
func HandleUserLogout(conn net.Conn, mid pb_common.Mid, msg interface{}) {
	packet := msg.(*pb.CLogout)
	fmt.Println("message handler mid:", mid, " body:", packet)

	name := packet.Name
	err := GetChatMgr().UserLogout(name)
	if err != nil {
		fmt.Printf("user logout error %s\n", err.Error())
		PushErrorMessage(conn, err)
		return
	}

	// 回复登出成功消息
	respMsg := &pb.SLogout{
		Name: name,
	}

	msgByte, err := proto.Marshal(respMsg)
	if err != nil {
		fmt.Printf("msg Marshal error %s\n", err.Error())
		return
	}

	SendPacket(conn, pb_common.Mid_G2C_USER_LOGOUT, msgByte)
}

// 处理消息发送
func HandleSendMessage(conn net.Conn, mid pb_common.Mid, msg interface{}) {
	packet := msg.(*pb.CSend)
	fmt.Println("message handler mid:", mid, " body:", packet)

	result, err := GetChatMgr().SendMessage(packet.Channel, packet.SenderName, packet.Content, false)
	if err != nil {
		fmt.Printf("send message err: %v\n", err)
		PushErrorMessage(conn, err)
		return
	}

	respMsg := &pb.SSend{
		Result: result,
	}

	msgByte, err := proto.Marshal(respMsg)
	if err != nil {
		fmt.Printf("msg Marshal error %s\n", err.Error())
		return
	}

	SendPacket(conn, pb_common.Mid_G2C_SEND_MESSAGE, msgByte)
}

// 处理获取聊天记录
func HandleGetChatLog(conn net.Conn, mid pb_common.Mid, msg interface{}) {
	packet := msg.(*pb.CGetLog)
	fmt.Println("message handler mid:", mid, " body:", packet)

	resp, err := GetChatMgr().GetChatLogs(packet.Channel)
	if err != nil {
		fmt.Printf("get chat log err: %v\n", err)
		PushErrorMessage(conn, err)
		return
	}

	respMsg := &pb.SGetLog{
		Logs: make([]*pb.ChatLog, 0),
	}

	for _, v := range resp.Logs {
		respMsg.Logs = append(respMsg.Logs, &pb.ChatLog{SenderName: v.SenderName, Content: v.Content, Timestamp: v.Timestamp})
	}

	msgByte, err := proto.Marshal(respMsg)
	if err != nil {
		fmt.Printf("msg Marshal error %s\n", err.Error())
		return
	}

	SendPacket(conn, pb_common.Mid_G2C_GET_LOGS, msgByte)
}

// 处理切换频道
func HandleChangeChannel(conn net.Conn, mid pb_common.Mid, msg interface{}) {
	packet := msg.(*pb.CChangeChannel)
	fmt.Println("message handler mid:", mid, " body:", packet)

	err := ChangeChannel(packet.Name, packet.Channel)
	if err != nil {
		fmt.Printf("user %s change channel failed, err: %v\n", packet.Name, err)
		PushErrorMessage(conn, err)
		return
	}

	resp, err := GetChatMgr().GetChatLogs(packet.Channel)
	if err != nil {
		fmt.Printf("get chat log err: %v\n", err)
		return
	}

	respMsg := &pb.SChangeChannel{
		Channel: packet.Channel,
		Logs:    make([]*pb.ChatLog, 0),
	}

	for _, v := range resp.Logs {
		respMsg.Logs = append(respMsg.Logs, &pb.ChatLog{SenderName: v.SenderName, Content: v.Content, Timestamp: v.Timestamp})
	}

	msgByte, err := proto.Marshal(respMsg)
	if err != nil {
		fmt.Printf("msg Marshal error %s\n", err.Error())
		return
	}

	SendPacket(conn, pb_common.Mid_G2C_CHANGE_CHANNEL, msgByte)

	GetChatMgr().ChangeChannel(packet.Name, packet.Channel)
}

// 处理获取频道列表
func HandleGetChannelList(conn net.Conn, mid pb_common.Mid, msg interface{}) {
	packet := msg.(*pb.CGetChannelList)
	fmt.Println("message handler mid:", mid, " body:", packet)

	resp, err := GetChatMgr().GetChannelList()
	if err != nil {
		fmt.Printf("get channel list err: %v\n", err)
		PushErrorMessage(conn, err)
		return
	}

	respMsg := &pb.SGetChannelList{
		List: make([]*pb.ChannelData, 0),
	}

	for _, v := range resp.List {
		respMsg.List = append(respMsg.List, &pb.ChannelData{Id: v.Id, Desc: v.Desc})
	}

	msgByte, err := proto.Marshal(respMsg)
	if err != nil {
		fmt.Printf("msg Marshal error %s\n", err.Error())
		return
	}

	SendPacket(conn, pb_common.Mid_G2C_GET_CHANNEL_LIST, msgByte)
}

// 推送错误消息给客户端
func PushErrorMessage(conn net.Conn, err error) {
	status, _ := status.FromError(err)
	respMsg := &pb_common.SErrorMessage{
		Code: int32(status.Code()),
		Msg:  status.Message(),
	}

	msgByte, err := proto.Marshal(respMsg)
	if err != nil {
		fmt.Printf("msg Marshal error %s\n", err.Error())
		return
	}

	SendPacket(conn, pb_common.Mid_G2C_ERROR_MESSAGE, msgByte)
}

// 发送消息包
func SendPacket(conn net.Conn, mid pb_common.Mid, msgByte []byte) {
	buf := &bytes.Buffer{}
	var head []byte
	head = make([]byte, 8)

	length := uint32(bytes.Count(msgByte, nil) - 1)
	binary.BigEndian.PutUint32(head[0:4], length)
	binary.BigEndian.PutUint32(head[4:8], uint32(mid))
	buf.Write(head[:8])
	buf.Write(msgByte)

	n, err := conn.Write(buf.Bytes())
	if err != nil {
		fmt.Printf("conn write failed, err = %v\n", err)
	} else {
		fmt.Printf("send packet mid: %d, len: %d, write length: %d\n", mid, length, n)
	}
}
