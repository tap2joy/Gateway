package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"go.elastic.co/apm/module/apmgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"

	"github.com/tap2joy/Gateway/server"
	"github.com/tap2joy/Gateway/services"
	"github.com/tap2joy/Gateway/utils"
	pb_common "github.com/tap2joy/Protocols/go/common"
	pb "github.com/tap2joy/Protocols/go/grpc/gateway"
)

func main() {
	// 初始化
	services.GetChatMgr().Init()

	go StartTcpServer()
	StartRpcServer()
}

func StartTcpServer() {
	lis, err := net.Listen("tcp", ":9108")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	fmt.Println("server is wating ....")
	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("connect failed ...")
		}
		fmt.Println(conn.RemoteAddr(), "connect success !")

		services.RegisterMsg(conn)
		go ServeHandle(conn)
	}
}

func StartRpcServer() {
	lis, err := net.Listen("tcp", ":9109")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		apmgrpc.NewUnaryServerInterceptor(apmgrpc.WithRecovery()),
		grpc_validator.UnaryServerInterceptor())))

	pb.RegisterGatewayServiceServer(s, &server.Server{})
	grpc_health_v1.RegisterHealthServer(s, &server.HealthServer{})
	reflection.Register(s)
	s.Serve(lis)
}

// 分包函数
func packetSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if !atEOF && len(data) >= 8 {
		var length int32
		// 读出 数据包中 实际数据 的长度
		binary.Read(bytes.NewReader(data[0:4]), binary.BigEndian, &length)
		packetLen := int(length) + 8
		dataLen := len(data)
		fmt.Printf("packetLen = %d, dataLen = %d\n", packetLen, dataLen)
		if packetLen <= dataLen {
			return packetLen, data[:packetLen], nil
		}
	}
	return
}

func ServeHandle(conn net.Conn) {
	for {
		readData := make([]byte, 64535)
		readLen, err := conn.Read(readData)

		if err != nil {
			if err == io.EOF {
				// 链接已关闭
				services.GetChatMgr().OnConnClosed(conn)
				fmt.Println("conn closed")
				break
			}
			continue
		}

		if readLen == 0 {
			fmt.Println("readLen == 0")
			continue
		}

		// 处理tcp粘包
		buf := bytes.NewBuffer(readData[:readLen])
		scanner := bufio.NewScanner(buf)
		scanner.Split(packetSplitFunc)

		for scanner.Scan() {
			packetBytes := scanner.Bytes()

			// 读取包头
			length := int32(0)
			var mid pb_common.Mid
			binary.Read(bytes.NewReader(packetBytes[0:4]), binary.BigEndian, &length)
			binary.Read(bytes.NewReader(packetBytes[4:8]), binary.BigEndian, &mid)

			if mid == pb_common.Mid_INVALID_MID || mid > 9999 {
				continue
			}

			fmt.Printf("receive packet mid: %d, len: %d\n", mid, length)

			services.HandleRawData(mid, length+8, packetBytes)
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("无效数据包")
			continue
		}
	}

	fmt.Println("conn end")
}

func InitService() {
	// 注册服务
	localAddress := utils.GetLocalAddress()
	err := services.RegisterGatewayService(localAddress)
	if err != nil {
		fmt.Printf("register gateway service failed, err = %v\n", err)
	} else {
		fmt.Println("register gateway service success")

		// 定时获取聊天服务列表，5秒一次
		utils.StartTimer(5, "2021-01-01 19:14:30", "", func() {
			services.GetChatServices()
		})
		select {}
	}
}
