package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	mea_gen_d "mea_go/api/mea.gen.d"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func protoToGo(protoImg *mea_gen_d.Image) *image.RGBA {
	w := int(protoImg.Info.Width)
	h := int(protoImg.Info.Height)
	total := w * h
	goImg := image.NewRGBA(image.Rect(0, 0, w, h))

	idxCalc := func(idx int) (int, int) {
		y := idx / w
		x := idx - w*y
		return x, y
	}
	var y int
	var x int
	var pixel []byte
	for i := range total {
		rgb_idx := i * 3
		pixel = protoImg.Pixels[rgb_idx : rgb_idx+3]
		c := color.RGBA{
			R: pixel[0],
			G: pixel[1],
			B: pixel[2],
			A: 255,
		}
		x, y = idxCalc(i)
		goImg.SetRGBA(x, y, c)
	}
	return goImg
}

func main() {
	// TODO: As we connecting with commpressed comfy api
	// we also consider retransimiting data over websocket,
	// for other technoogies not supported by main protocol
	// but at first it can be simple cli?! inpaint editor??

	prompts := []string{
		"warrior wakes up early morning, in his friend house after yesterdey's happy bonfire day:D",
		"mage wakes up erlier so he can go fishing with uncle bob",
	}

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

	rpc_stub := mea_gen_d.NewComfyClient(conn)
	fmt.Println(prompts[1:])

	a := mea_gen_d.Options{
		Prompts:  []string{""},
		ImgPower: 1,
		Seed:     0,
		InptFlag: mea_gen_d.InpaintType_SDXL,
	}

	for i := range len(prompts) {
		a.Prompts = prompts[i : i+1]
		rpc_stub.SetOptions(context.Background(), &a)
		fmt.Println("+++ gen prompt: ", a.Prompts)
		protoImg, err := rpc_stub.Txt2Img(context.Background(), &mea_gen_d.Empty{})
		if err != nil {
			log.Fatal(err.Error())
		}

		goImg := protoToGo(protoImg)

		pngFile := fmt.Sprintf("fs/img_%d.png", i)
		file, err := os.Create(pngFile)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = png.Encode(file, goImg)
		if err != nil {
			log.Fatal(err.Error())
		}
		file.Close()
		fmt.Println("+++ image saved: ", pngFile)
	}
}
