package main

import (
	"fmt"

	"github.com/lsq51201314/anythingllm-stream-chat/v1"
)

func main() {
	hp := anythingllm.New("192.168.100.123", 3001, "pw", "AVYS692-YQH45W3-N5AHZAX-39TPS7V")
	hp.SetChatCallback(chatCallback)
	hp.StreamChat("你好啊，这是一个测试的聊天，请随便回复点什么吧！")
}

func chatCallback(uuid string, message string) {
	fmt.Print(message)
}
