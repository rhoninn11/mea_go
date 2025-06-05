package internal

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"log"
	mea_gen_d "mea_go/api/mea.gen.d"
	"mea_go/components"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/a-h/templ"
)

var HeaderContentType = "Content-Type"

const (
	ContentTypeHtml        = "text/html"
	ContentTypeEventStream = "text/event-stream"
	ContentTypePng         = "image/png"
)

const feedId = "feedID"

type HttpFuncMap = map[templ.SafeURL]HttpFunc

type PromptMap = map[string]string
type ImgMap = map[string][]byte

type GenState struct {
	prompts     PromptMap
	promptSlots []string
	comfyData   *ComfyData
	imageData   ImgMap
}

var memory GenState

func init() {
	memory.init()
	fmt.Println("+++ prompt module inited")
}

func (gs *GenState) init() {
	gs.promptSlots = []string{
		"slot_a",
		"slot_b",
		"slot_c",
	}
	size := len(gs.promptSlots)
	gs.prompts = make(PromptMap, size)
	gs.imageData = make(ImgMap, size)
	for _, slot := range memory.promptSlots {
		gs.prompts[slot] = "placeholder"
	}

	imgDir := "_fs/img"
	entries, err := os.ReadDir(imgDir)
	if err != nil {
		log.Fatalln("scaning imgs", err.Error())
	}

	var errs = make([]error, 0, 64)
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".png") {
			continue
		}

		id := strings.TrimSuffix(name, ".png")
		imgPath := strings.Join([]string{imgDir, name}, "/")

		data, err := os.ReadFile(imgPath)
		if err != nil {
			err = fmt.Errorf("!!! failed at %s, %v", id, err)
			errs = append(errs, err)
		}
		log.Default().Println("INFO: ", imgPath, " loaded")
		gs.imageData[id] = data
	}
	for _, err := range errs {
		log.Default().Println(err.Error())
		defer log.Fatal()
	}
}

func (ps *GenState) setPrompt(slot string, text string) {
	if _, exist := ps.prompts[slot]; exist {
		ps.prompts[slot] = text
	}
}

func PromptEditor() templ.Component {
	// calls := "/prompt"
	targetID := fmt.Sprintf("#%s", feedId)
	edits := []templ.Component{
		components.PromptPad("slot_a", targetID),
		components.PromptPad("slot_b", targetID),
		components.PromptPad("slot_c", targetID),
		components.GenButton(targetID),
	}
	return components.FeedColumn(edits, feedId)
}

func (ps *GenState) GenPage(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		if r.ParseForm() != nil {
			http.Error(w, "!!! not specified", 500)
		}

		for slot, v := range r.Form {
			ps.setPrompt(slot, strings.Join(v, ""))
			fmt.Printf("+++ slot: %s, updated\n", slot)
		}
	}

	w.Header().Set(HeaderContentType, ContentTypeHtml)

	imgs := make([]templ.Component, 0, len(ps.imageData)+1)
	for k, _ := range ps.imageData {
		imgComp := components.JustImg(imgUrl(k))
		imgs = append(imgs, imgComp)
	}

	imgs = append(imgs, PromptEditor())
	feed := components.FeedColumn(imgs, "imgs")

	feed.Render(context.Background(), w)
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
	gen.imageData[imgName] = buffer.Bytes()

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

func (ps *GenState) PromptCommit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "!!! commit on get", 500)
	}

	for k, v := range ps.prompts {
		fmt.Println("+++", k, v)
	}
	id, err := imageGen(ps, ps.comfyData)
	if err != nil {
		log.Default().Println("ERR: ", err.Error())
		http.Error(w, "", 500)
		return
	}

	w.Header().Set(HeaderContentType, ContentTypeHtml)
	feed := components.FeedColumn(
		[]templ.Component{
			components.JustImg(imgUrl(id)),
			PromptEditor(),
		}, "xd")
	feed.Render(context.Background(), w)
}

var img []byte

func loadImage() ([]byte, error) {
	bData, err := os.ReadFile("fs/image.png")

	if err != nil {
		return nil, err
	}
	return bData, nil
}

func (ps *GenState) FetchImage(w http.ResponseWriter, r *http.Request) {

	vals := r.URL.Query()
	id, ok := vals["id"]
	if !ok {
		http.Error(w, "bad request", 500)
		return
	}

	imgData, ok := ps.imageData[id[0]]
	if !ok {
		http.Error(w, "bad request", 500)
	}

	w.Header().Set(HeaderContentType, ContentTypePng)
	size, err := w.Write(imgData)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Println("+++ sended ", size, " bytes")
}

func PromptModuleAccess() *GenState {

	defaultComfy := SpawComfyDefault()
	memory.comfyData = &defaultComfy
	return &memory
}
func imgUrl(id string) string {
	return fmt.Sprintf("/prompt/img?id=%s", id)
}

func (gs *GenState) LoadFns() HttpFuncMap {
	return HttpFuncMap{
		"/prompt":        gs.GenPage,
		"/prompt/commit": gs.PromptCommit,
		"/prompt/img":    gs.FetchImage,
	}
}
