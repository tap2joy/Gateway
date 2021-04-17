package test

import (
	"fmt"
	"testing"

	"github.com/tap2joy/Gateway/services"
)

func TestGetChannelList(t *testing.T) {
	resp, err := services.GetChatMgr().GetChannelList()
	if err != nil {
		fmt.Printf("get channel list err: %v\n", err)
	} else {
		for _, v := range resp.List {
			fmt.Printf("%v\n", v)
		}
	}
}

func TestGetChatLog(t *testing.T) {
	resp, err := services.GetChatMgr().GetChatLogs(1)
	if err != nil {
		fmt.Printf("get chat log err: %v\n", err)
	} else {
		for _, v := range resp.Logs {
			fmt.Printf("%v\n", v)
		}
	}
}
