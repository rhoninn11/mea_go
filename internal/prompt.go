package internal

import (
	"context"
	"fmt"
	"log"
	"math"
	"mea_go/components"
	"net/http"
	"os"
	"strings"

	"github.com/a-h/templ"
)

const (
	HContentType  = "Content-Type"
	HCacheControl = "Cache-Control"
)

const (
	ContentTypePlainText   = "text/plain"
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
	imageIds    []string
}

func (gs *GenState) addImage(id string, data []byte) {
	gs.imageIds = append(gs.imageIds, id)
	gs.imageData[id] = data
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
	gs.imageData = make(ImgMap, 128)
	gs.imageIds = make([]string, 0, 128)
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
		gs.imageIds = append(gs.imageIds, id)
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

func (ps *GenState) PromptInput(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		if r.ParseForm() != nil {
			http.Error(w, "!!! not specified", 500)
		}

		for slot, v := range r.Form {
			ps.setPrompt(slot, strings.Join(v, ""))
			fmt.Printf("+++ slot: %s, updated\n", slot)
		}
	}

	w.WriteHeader(200)
}

// show all results and editor
func (ps *GenState) GenPage(w http.ResponseWriter, r *http.Request) {

	// if r.Method == http.MethodPost {
	// 	if r.ParseForm() != nil {
	// 		http.Error(w, "!!! not specified", 500)
	// 	}

	// 	for slot, v := range r.Form {
	// 		ps.setPrompt(slot, strings.Join(v, ""))
	// 		fmt.Printf("+++ slot: %s, updated\n", slot)
	// 	}
	// }

	w.Header().Set(HContentType, ContentTypeHtml)

	const colNum = 4
	var imgNum = len(ps.imageIds)
	var rowNum = int(math.Ceil(float64(imgNum) / colNum))
	// fmt.Println("+++ rows ", rowNum, " cols ", colNum)

	rows := make([]templ.Component, 0, rowNum)
	row := make([]templ.Component, colNum)

	lastElem := colNum - 1
	var showedImgNum = 0
	for i, id := range ps.imageIds {
		imgComp := components.JustImg(imgUrl(id))
		elemIdx := i % colNum
		row[elemIdx] = imgComp
		if elemIdx == lastElem {
			rowCopy := make([]templ.Component, colNum)
			copy(rowCopy, row)
			rows = append(rows, components.FlexRow(rowCopy))
			showedImgNum += len(rowCopy)
		}
	}
	delta := imgNum - showedImgNum
	if delta > 0 {
		rows = append(rows, components.FlexRow(row[0:delta]))
	}

	rows = append(rows, PromptEditor())
	feed := components.FeedColumn(rows, "imgs")
	wholePage := PageWithSidebar(feed)
	wholePage.Render(context.Background(), w)
}

// Slightly drunk medevil gardener is suprised that his bell become so big
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

	w.Header().Set(HContentType, ContentTypeHtml)
	feed := components.FeedColumn(
		[]templ.Component{
			components.JustImg(imgUrl(id)),
			PromptEditor(),
		}, "xd")
	feed.Render(context.Background(), w)
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

	w.Header().Set(HContentType, ContentTypePng)
	_, err := w.Write(imgData)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
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
		"/gen_page":      gs.GenPage,
		"/prompt":        gs.PromptInput,
		"/prompt/commit": gs.PromptCommit,
		"/prompt/img":    gs.FetchImage,
	}
}
