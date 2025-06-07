package internal

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"log"
	"os"
	"time"

	mea_gen_d "mea_go/api/mea.gen.d"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var baseOptions mea_gen_d.Options

func init() {
	baseOptions = mea_gen_d.Options{
		Prompts:  []string{""},
		ImgPower: 1,
		Seed:     0,
		InptFlag: mea_gen_d.InpaintType_SDXL,
	}
}

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

func SpawComfyDefault() ComfyData {
	var port = 50051
	var host = "0.0.0.0"

	baseOptions = mea_gen_d.Options{
		Prompts:  []string{""},
		ImgPower: 1,
		Seed:     0,
		InptFlag: mea_gen_d.InpaintType_SDXL,
	}

	return ComfyData{
		Service: ConnectComfy(host, port),
		Options: &baseOptions,
		Ctx:     context.Background(),
	}
}
func uniqueName() string {
	timestump := time.Now().UTC().UnixMilli()
	return fmt.Sprintf("%d", timestump)
}

func imageGen(gen *GenState, comfy *ComfyData) (string, error) {
	var _plug mea_gen_d.Empty
	firsSlot := gen.promptSlots[0]
	prompt := gen.prompts[firsSlot]

	opt := comfy.Options
	serv := comfy.Service

	opt.Prompts = []string{prompt}
	if _, err := serv.SetOptions(comfy.Ctx, opt); err != nil {
		return "", fmt.Errorf("!!! options failed, %v", err)
	}

	pImg, err := serv.Txt2Img(comfy.Ctx, &_plug)
	if err != nil {
		return "", fmt.Errorf("!!! txt2img failed, %v", err)
	}

	imgName := uniqueName()
	gImg := ImgProtoToGo(pImg)
	var buffer = bytes.Buffer{}
	if err := png.Encode(&buffer, gImg); err != nil {
		return "", fmt.Errorf("!!! failed to encode %s, %v", imgName, err)
	}
	gen.addImage(imgName, buffer.Bytes())

	fileName := fmt.Sprintf("_fs/img/%s.png", imgName)
	file, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf("!!! file failed to open %s, %v", fileName, err)
	}
	defer file.Close()

	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return "", fmt.Errorf("!!! file write fail %s, %v", fileName, err)
	}
	return imgName, nil

}

func (gd *ComfyData) GenFew(prompts []string) error {

	comfy := gd.Service
	for i := range len(prompts) {
		baseOptions.Prompts = prompts[i : i+1]
		comfy.SetOptions(context.Background(), &baseOptions)
		fmt.Println("+++ gen prompt: ", baseOptions.Prompts)
		protoImg, err := comfy.Txt2Img(context.Background(), &mea_gen_d.Empty{})
		if err != nil {
			return fmt.Errorf("!!! request failed")
		}

		goImg := ImgProtoToGo(protoImg)
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
