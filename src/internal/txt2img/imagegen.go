package txt2img

import (
	"bytes"
	"fmt"
	"image/png"
	mea_gen_d "mea_go/src/api/mea.gen.d"
	utils "mea_go/src/internal"
	"os"
	"time"
)

func uniqueName() string {
	timestump := time.Now().UTC().UnixMilli()
	return fmt.Sprintf("%d", timestump)
}

type SlotPrompt struct {
	SlotName string `yaml:"slot"`
	Prompt   string `yaml:"name"`
}
type SlotPromptS struct {
	Promps   []SlotPrompt `yaml:"prompts"`
	Sequence []string     `yaml:"seq"`
}

func FormPrompt(usedSlots []string, usedPrompts []string) SlotPromptS {
	var few = make([]SlotPrompt, 0, len(usedSlots))
	for i := range usedSlots {
		single := SlotPrompt{
			SlotName: usedSlots[i],
			Prompt:   usedPrompts[i],
		}
		few = append(few, single)
	}

	return SlotPromptS{
		Promps:   few,
		Sequence: usedSlots,
	}
}
func ImageGen(gen *GenState, comfy *ComfyData) (string, error) {
	var _plug mea_gen_d.Empty
	var imgBasename = uniqueName()

	opt := comfy.Options
	serv := comfy.Service

	usedSlots := make([]string, 0, 4)
	usedPrompts := make([]string, 0, 4)
	for name, slot := range SlotMapping {
		sloted := mea_gen_d.SlotedPrompt{
			Slot:   slot,
			Prompt: gen.prompts[slot],
		}
		if sloted.Prompt == "" {
			continue
		}
		usedSlots = append(usedSlots, name)
		usedPrompts = append(usedPrompts, sloted.Prompt)
		if _, err := serv.SetPrompt(comfy.Ctx, &sloted); err != nil {
			return "", fmt.Errorf("failed to set prompt (%s) | %v", name, err)
		}
	}

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

	gImg := utils.ImgProtoToGo(pImg)
	var buffer = bytes.Buffer{}
	if err := png.Encode(&buffer, gImg); err != nil {
		return "", fmt.Errorf("!!! failed to encode %s, %v", imgBasename, err)
	}

	//updating state
	gen.addImage(imgBasename, buffer.Bytes())

	//saving image
	dirImg := utils.DirImage()
	pngFile := utils.JoinPath(dirImg, utils.PngFilename(imgBasename))
	if err := data2File(pngFile, buffer); err != nil {
		return "", fmt.Errorf("!!! failed to encode %s, %v", imgBasename, err)
	}

	yamlFile := utils.JoinPath(dirImg, utils.YamlFilename(imgBasename))
	yamlObj := FormPrompt(usedSlots, usedPrompts)
	err = utils.SaveAsYAML(yamlFile, yamlObj)
	if err != nil {
		return "", fmt.Errorf("!!! failed to save %s, %w", yamlFile, err)
	}

	return imgBasename, nil
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
