package services

import (
	"errors"
	"net"
	"reflect"

	"github.com/golang/protobuf/proto"

	pb_common "github.com/tap2joy/Protocols/go/common"
	pb "github.com/tap2joy/Protocols/go/gateway"
)

type MessageHandler func(conn net.Conn, mid pb_common.Mid, msg interface{})

type MessageInfo struct {
	MsgType    reflect.Type
	MsgHandler MessageHandler
	Conn       net.Conn
}

var (
	MsgMap = make(map[pb_common.Mid]MessageInfo)
)

func RegisterMessage(conn net.Conn, mid pb_common.Mid, msg interface{}, handler MessageHandler) {
	var info MessageInfo
	info.MsgType = reflect.TypeOf(msg.(proto.Message))
	info.MsgHandler = handler
	info.Conn = conn

	MsgMap[mid] = info
}

// 注册消息处理
func RegisterMsg(conn net.Conn) {
	RegisterMessage(conn, pb_common.Mid_C2G_USER_LOGIN, &pb.CLogin{}, HandleUserLogin)
	RegisterMessage(conn, pb_common.Mid_C2G_USER_LOGOUT, &pb.CLogout{}, HandleUserLogout)
	RegisterMessage(conn, pb_common.Mid_C2G_SEND_MESSAGE, &pb.CSend{}, HandleSendMessage)
	RegisterMessage(conn, pb_common.Mid_C2G_GET_LOGS, &pb.CGetLog{}, HandleGetChatLog)
	RegisterMessage(conn, pb_common.Mid_C2G_CHANGE_CHANNEL, &pb.CChangeChannel{}, HandleChangeChannel)
	RegisterMessage(conn, pb_common.Mid_C2G_GET_CHANNEL_LIST, &pb.CGetChannelList{}, HandleGetChannelList)
}

func HandleRawData(mid pb_common.Mid, length int32, data []byte) error {
	if info, ok := MsgMap[mid]; ok {
		msg := reflect.New(info.MsgType.Elem()).Interface()
		err := proto.Unmarshal(data[8:length], msg.(proto.Message))
		if err != nil {
			return err
		}
		info.MsgHandler(info.Conn, mid, msg)
		return err
	}
	return errors.New("not found mid")
}
