package txt2img

import (
	"bytes"
	"fmt"
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

func (sps *SlotPromptS) init(chain []mea_gen_d.SlotedPrompt) SlotPromptS {
	return FormPrompt(chain)
}

func FormPrompt(chain []mea_gen_d.SlotedPrompt) SlotPromptS {
	well := make([]SlotPrompt, 0, len(chain))
	sequence := make([]string, 0, len(chain))
	for _, one := range chain {
		var slotName = one.Slot.String()
		single := SlotPrompt{
			SlotName: slotName,
			Prompt:   one.Prompt,
		}
		well = append(well, single)
		sequence = append(sequence, slotName)
	}

	return SlotPromptS{
		Promps:   well,
		Sequence: sequence,
	}
}
func ImageGen(gen *GenState, comfy *ComfyData) (string, error) {
	var _plug mea_gen_d.Empty
	var imgBasename = uniqueName()

	comfyOpts := comfy.Options
	comfyGrpc := comfy.Service

	var slots = []mea_gen_d.Slot{mea_gen_d.Slot_a, mea_gen_d.Slot_b, mea_gen_d.Slot_c}
	var usedSlots = make([]mea_gen_d.Slot, 0, len(slots))
	var slotedPrompts = make([]mea_gen_d.SlotedPrompt, 0, len(slots))

	for _, slot := range slots {
		prompt := gen.prompts[slot]
		if prompt == "" {
			continue
		}

		sloted := mea_gen_d.SlotedPrompt{
			Slot:   slot,
			Prompt: prompt,
		}
		slotedPrompts = append(slotedPrompts, sloted)
		usedSlots = append(usedSlots, slot)
		if _, err := comfyGrpc.SetPrompt(comfy.Ctx, &sloted); err != nil {
			return "", fmt.Errorf("failed to set prompt (%s) | %v", slot.String(), err)
		}
	}

	comfyOpts.PromptChain = usedSlots

	if _, err := comfyGrpc.SetOptions(comfy.Ctx, comfyOpts); err != nil {
		return "", fmt.Errorf("!!! options failed, %w", err)
	}

	pImg, err := comfyGrpc.Txt2Img(comfy.Ctx, &_plug)
	if err != nil {
		return "", fmt.Errorf("!!! txt2img failed, %w", err)
	}

	yamlObj := FormPrompt(slotedPrompts)
	gImg := utils.ImgProtoToGo(pImg)

	//updating state
	err = gen.addImage(imgBasename, gImg, yamlObj)
	if err != nil {
		return "", fmt.Errorf("img failed to save | %w", err)
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
