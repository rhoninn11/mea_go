package internal

import (
	"bytes"
	"fmt"
	"image/png"
	mea_gen_d "mea_go/src/api/mea.gen.d"
	"os"
	"slices"
	"time"
)

func uniqueName() string {
	timestump := time.Now().UTC().UnixMilli()
	return fmt.Sprintf("%d", timestump)
}
func imageGen(gen *GenState, comfy *ComfyData) (string, error) {
	var _plug mea_gen_d.Empty

	opt := comfy.Options
	serv := comfy.Service

	usedPrompts := make([]string, 0, 4)
	for name, slot := range SlotMapping {
		slotedPrompt := mea_gen_d.SlotedPrompt{
			Slot:   slot,
			Prompt: gen.prompts[slot],
		}
		if slotedPrompt.Prompt == "" {
			continue
		}
		usedPrompts = append(usedPrompts, name)
		if _, err := serv.SetPrompt(comfy.Ctx, &slotedPrompt); err != nil {
			return "", fmt.Errorf("failed to set prompt (%s) | %v", name, err)
		}
	}
	slices.Sort(usedPrompts)
	fmt.Println("+++ used prompts: ", usedPrompts)

	// firsSlot := gen.promptSlots[0]
	// prompt := gen.prompts[firsSlot]

	// prompcik := mea_gen_d.SlotedPrompt{
	// 	Slot:   mea_gen_d.Slot_a,
	// 	Prompt: prompt,
	// }
	// if _, err := serv.SetPrompt(comfy.Ctx, &prompcik); err != nil {
	// 	return "", fmt.Errorf("failed to set prompt | %v", err)
	// }

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

	//updating state
	gen.addImage(imgName, buffer.Bytes())

	//saving image
	pngFile := JoinPath(DirImage(), PngFilename(imgName))
	if err := data2File(pngFile, buffer); err != nil {
		return "", fmt.Errorf("!!! failed to encode %s, %v", imgName, err)
	}

	//TODO: saving prompts
	yamlFile := JoinPath(DirImage(), YamlFilename(imgName))
	_ = yamlFile

	return imgName, nil
}

func data2File(fileName string, data bytes.Buffer) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("!!! file failed to open %s, %v", fileName, err)
	}
	defer file.Close()

	_, err = file.Write(data.Bytes())
	if err != nil {
		return fmt.Errorf("!!! file write fail %s, %v", fileName, err)
	}

	return nil
}
