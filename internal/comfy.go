package internal

import (
	"context"
	"fmt"
	"image/png"
	"log"
	"os"

	mea_gen_d "mea_go/api/mea.gen.d"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectComfy(host string, port int) *mea_gen_d.ComfyClient {
	// caFile := "assets/root.crt"
	// credentials.NewClientTLSFromFile(caFile, "")

	// default gRPC???
	serv_address := fmt.Sprintf("%s:%d", host, port)

	var opts = grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(serv_address, opts)
	if err != nil {
		log.Fatal(fmt.Errorf("!!! ^at new client %w", err))
	}

	comfyApi := mea_gen_d.NewComfyClient(conn)
	return &comfyApi
}

type ComfyData struct {
	ComfyClient *mea_gen_d.ComfyClient
}

func SpawComfyDefault() ComfyData {
	var port = 50051
	var host = "0.0.0.0"

	return ComfyData{
		ComfyClient: ConnectComfy(host, port),
	}
}

func (gd *ComfyData) GenFew(prompts []string) error {
	a := mea_gen_d.Options{
		Prompts:  []string{""},
		ImgPower: 1,
		Seed:     0,
		InptFlag: mea_gen_d.InpaintType_SDXL,
	}

	comfy := *gd.ComfyClient
	for i := range len(prompts) {
		a.Prompts = prompts[i : i+1]
		comfy.SetOptions(context.Background(), &a)
		fmt.Println("+++ gen prompt: ", a.Prompts)
		protoImg, err := comfy.Txt2Img(context.Background(), &mea_gen_d.Empty{})
		if err != nil {
			return fmt.Errorf("!!! request failed")
		}

		goImg := ProtoToGo(protoImg)
		pngFile := fmt.Sprintf("fs/img_%d.png", i)
		file, err := os.Create(pngFile)
		if err != nil {
			return fmt.Errorf("!!! failed to create %s", pngFile)
		}
		if err := png.Encode(file, goImg); err != nil {
			return fmt.Errorf("!!! failed to encode %s", pngFile)
		}
		file.Close()
		fmt.Println("+++ image saved: ", pngFile)
	}

	return nil
}
