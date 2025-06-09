package internal

import (
	"bytes"
	"fmt"
	"image/png"
	mea_gen_d "mea_go/api/mea.gen.d"
	"os"
	"time"
)

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
