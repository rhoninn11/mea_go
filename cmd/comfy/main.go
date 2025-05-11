package main

import (
	"context"
	"fmt"
	"log"

	mea_gen_d "mea_go/api/mea.gen.d"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	var opts []grpc.DialOption

	caFile := "assets/root.crt"
	credentials.NewClientTLSFromFile(caFile, "")

	const port = 50051
	serv_address := fmt.Sprintf("0.0.0.0:%d", port)
	conn, err := grpc.NewClient(serv_address, opts...)
	if err != nil {
		log.Fatal(err.Error())
	}

	hmm := mea_gen_d.NewComfyClient(conn)

	a := mea_gen_d.Options{
		Prompts:  prompts,
		ImgPower: 1,
		Seed:     0,
		InptFlag: mea_gen_d.InpaintType_SDXL,
	}

	hmm.SetOptions(context.Background(), &a)
}
