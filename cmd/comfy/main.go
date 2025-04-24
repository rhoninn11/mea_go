package main

import (
	"fmt"
)

func main() {
	todoMeassage := "TODO: here will be implemented client calling comfyui over grpc\n" +
		"Maybe explor posibility of usage jsonrpc, or sending proto buffers over websocket"
	fmt.Println(todoMeassage)

	//TODO: parę zdjęć można by tu wygenerować

	prompts := []string{
		"warrior wakes up early morning, in his friend house after yesterdey's happy bonfire day:D",
		"mage wakes up erlier so he can go fishing with uncle bob",
	}
}
