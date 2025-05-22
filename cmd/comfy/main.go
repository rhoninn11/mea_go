package main

import (
	"context"
	"fmt"
	"image/png"
	"log"
	"os"

	mea_gen_d "mea_go/api/mea.gen.d"
	"mea_go/internal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func connectComfy() *mea_gen_d.ComfyClient {
	// caFile := "assets/root.crt"
	// credentials.NewClientTLSFromFile(caFile, "")

	// default gRPC???
	const port = 50051
	serv_address := fmt.Sprintf("0.0.0.0:%d", port)

	var opts = grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(serv_address, opts)
	if err != nil {
		log.Fatal(fmt.Errorf("^at new client %w", err))
	}

	comfyApi := mea_gen_d.NewComfyClient(conn)
	return &comfyApi
}

type GlobData struct {
	comfy *mea_gen_d.ComfyClient
}

func (gd *GlobData) genFew() error {
	prompts := []string{
		"warrior wakes up early morning, in his friend house after yesterdey's happy bonfire day:D",
		"mage wakes up erlier so he can go fishing with uncle bob",
	}
	a := mea_gen_d.Options{
		Prompts:  []string{""},
		ImgPower: 1,
		Seed:     0,
		InptFlag: mea_gen_d.InpaintType_SDXL,
	}

	comfy := *gd.comfy
	for i := range len(prompts) {
		a.Prompts = prompts[i : i+1]
		comfy.SetOptions(context.Background(), &a)
		fmt.Println("+++ gen prompt: ", a.Prompts)
		protoImg, err := comfy.Txt2Img(context.Background(), &mea_gen_d.Empty{})
		if err != nil {
			return fmt.Errorf("!!! request failed")
		}

		goImg := internal.ProtoToGo(protoImg)
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

func main() {
	// TODO: As we connecting with commpressed comfy api
	// we also consider retransimiting data over websocket,
	// for other technoogies not supported by main protocol
	// but at first it can be simple cli?! inpaint editor??

	glob := GlobData{
		comfy: connectComfy(),
	}

	glob.genFew()

}
