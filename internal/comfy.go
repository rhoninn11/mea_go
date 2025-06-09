package internal

import (
	"context"
	"fmt"
	"log"

	mea_gen_d "mea_go/api/mea.gen.d"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectComfy(host string, port int) mea_gen_d.ComfyClient {
	// caFile := "assets/root.crt"
	// credentials.NewClientTLSFromFile(caFile, "")

	// default gRPC???
	serv_address := fmt.Sprintf("%s:%d", host, port)

	var opts = grpc.WithTransportCredentials(insecure.NewCredentials())
	fmt.Printf("+++ connecting to comfy %s:%d\n", host, port)
	conn, err := grpc.NewClient(serv_address, opts)
	if err != nil {
		log.Fatal(fmt.Errorf("!!! ^at new client %w", err))
	}

	comfyApi := mea_gen_d.NewComfyClient(conn)
	return comfyApi
}

type ComfyData struct {
	Service mea_gen_d.ComfyClient
	Options *mea_gen_d.Options
	Ctx     context.Context
}

func DefaultOpts() mea_gen_d.Options {
	return mea_gen_d.Options{
		Prompts:  []string{""},
		ImgPower: 1,
		Seed:     0,
		InptFlag: mea_gen_d.InpaintType_SDXL,
	}
}

func DefaultComfySpawn() ComfyData {
	var port = 50051
	var host = "0.0.0.0"

	baseOptions := DefaultOpts()
	return ComfyData{
		Service: ConnectComfy(host, port),
		Options: &baseOptions,
		Ctx:     context.Background(),
	}
}
