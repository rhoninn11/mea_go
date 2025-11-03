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

const feedId = "feedID"

const (
	SLOT_A = "slot_a"
	SLOT_B = "slot_b"
	SLOT_C = "slot_c"
)

type HttpFuncMap = map[templ.SafeURL]HttpFunc

type PromptMap = map[string]string
type ImgMap = map[string][]byte

type PromptSlots = []string

type GenState struct {
	prompts     PromptMap
	promptSlots []string
	comfyData   *ComfyData
	imageData   ImgMap
	imageIds    []string
}

type FlowData struct {
	slots PromptSlots
	rowID string
	colID string
}

// hmm maybe simple squil would be adequate, like SQLite or Postgress
func (gs *GenState) addImage(id string, data []byte) {
	gs.imageIds = append(gs.imageIds, id)
	gs.imageData[id] = data
}

var memory GenState

func init() {
	memory.init()
	memory.fsImageFetch()
}

func (gs *GenState) fsImageFetch() {
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

func (gs *GenState) init() {
	gs.promptSlots = []string{SLOT_A, SLOT_B, SLOT_C}

	size := len(gs.promptSlots)
	gs.prompts = make(PromptMap, size)
	gs.imageData = make(ImgMap, 128)
	gs.imageIds = make([]string, 0, 128)
	for _, slot := range memory.promptSlots {
		gs.prompts[slot] = "placeholder"
	}
}

func PromptEditor(editorID string) templ.Component {
	// calls := "/prompt"
	submmitBtn := components.GenButton(fmt.Sprintf("#%s", editorID))
	editor := []templ.Component{
		components.PromptPad(SLOT_A, SLOT_A),
		components.PromptPad(SLOT_B, SLOT_B),
		components.PromptPad(SLOT_C, SLOT_C),
		submmitBtn,
	}
	return components.FeedColumn(editor, editorID)
}

func (ps *GenState) PromptInput(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		if r.ParseForm() != nil {
			http.Error(w, "!!! not specified", 500)
			return
		}

		for slot, data := range r.Form {
			prompt := strings.Join(data, "")
			if _, ok := ps.prompts[slot]; ok {
				ps.prompts[slot] = prompt
			} else {
				http.Error(w, "!!! not specified", 500)
				return
			}
		}
	}
	w.WriteHeader(200)
}

// show all results and editor
func (ps *GenState) GenPage(w http.ResponseWriter, r *http.Request) {
	SetContentType(w, ContentTypeHtml)

	const colNum = 4
	var imgNum = len(ps.imageIds)
	var rowNum = int(math.Ceil(float64(imgNum) / colNum))
	// fmt.Println("+++ rows ", rowNum, " cols ", colNum)

	rows := make([]templ.Component, 0, rowNum)
	row := make([]templ.Component, colNum)

	lastElem := colNum - 1
	var showedImgNum = 0
	for i, id := range ps.imageIds {
		imgComp := components.JustImg(imgUrl(id), imgDelUrl(id))
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

	rows = append(rows, PromptEditor("prompt_editor"))
	feed := components.FeedColumn(rows, "imgs")
	wholePage := PageWithSidebar(feed)
	wholePage.Render(context.Background(), w)
}

func (ps *GenState) PromptCommit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "!!! commit on get", 500)
	}

	for k, v := range ps.prompts {
		fmt.Println("+++ kv:", k, v)
	}

	log.Fatal("exit:D")
	// tu bym mógł zapisać prompty do bazy danych
	// może jako proto jako blob binarny w sql lite?
	id, err := imageGen(ps, ps.comfyData)
	if err != nil {
		log.Default().Println("ERR: ", err.Error())
		http.Error(w, "", 500)
		return
	}

	SetContentType(w, ContentTypeHtml)
	feed := components.FeedColumn(
		[]templ.Component{
			components.JustImg(imgUrl(id), imgDelUrl(id)),
			PromptEditor("prompt_editor"),
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

	SetContentType(w, ContentTypeHtml)
	_, err := w.Write(imgData)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
func (ps *GenState) DeleteImage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	w.WriteHeader(200)
	fmt.Println("+++ delete call for ", id)
}

func PromptModuleAccess() *GenState {

	defaultComfy := DefaultComfySpawn()
	memory.comfyData = &defaultComfy
	return &memory
}
func imgUrl(id string) string {
	return fmt.Sprintf("/prompt/img?id=%s", id)
}
func imgDelUrl(id string) string {
	return fmt.Sprintf("/prompt/img/del/%s", id)
}

func (gs *GenState) LoadFns() HttpFuncMap {
	return HttpFuncMap{
		"/gen_page":            gs.GenPage,
		"/prompt":              gs.PromptInput,
		"/prompt/commit":       gs.PromptCommit,
		"/prompt/img":          gs.FetchImage,
		"/prompt/img/del/{id}": gs.DeleteImage,
	}
}
