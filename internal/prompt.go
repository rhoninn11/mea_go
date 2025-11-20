package internal

import (
	"context"
	"fmt"
	"log"
	"math"
	"mea_go/components"
	"net/http"
	"os"
	"slices"
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

var memory GenState

func init() {
	memory.init()
}

// hmm maybe simple squil would be adequate, like SQLite or Postgress
func (gs *GenState) addImage(id string, data []byte) {
	gs.imageIds = append(gs.imageIds, id)
	gs.imageData[id] = data
}

func (gs *GenState) init() {
	gs.promptSlots = []string{SLOT_A, SLOT_B, SLOT_C}

	size := len(gs.promptSlots)
	gs.prompts = make(PromptMap, size)
	gs.imageData = make(ImgMap, 128)
	gs.imageIds = make([]string, 0, 128)
	for _, slot := range memory.promptSlots {
		gs.prompts[slot] = ""
	}

	func(gs *GenState) {
		imgDir := "_fs/img"
		entries, err := os.ReadDir(imgDir)
		if err != nil {
			log.Fatalln("scaning imgs", err.Error())
		}
		panicker := Panicker(4)
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".png") {
				continue
			}

			id := strings.TrimSuffix(name, ".png")
			imgPath := strings.Join([]string{imgDir, name}, "/")

			data, err := os.ReadFile(imgPath)
			if panicker.HasError(err) {
				continue
			}

			log.Default().Println("INFO: ", imgPath, " loaded")
			gs.imageIds = append(gs.imageIds, id)
			gs.imageData[id] = data
		}
	}(gs)
}

func (gs *GenState) PromptEditor(editorID string) templ.Component {
	// calls := "/prompt"

	padFromSlot := func(id string, slot string) templ.Component {
		return components.PromptPad(id, slot, gs.prompts[slot])
	}
	submmitBtn := components.GenButton(fmt.Sprintf("#%s", editorID))

	editor := []templ.Component{
		padFromSlot(SLOT_A, SLOT_A),
		padFromSlot(SLOT_B, SLOT_B),
		padFromSlot(SLOT_C, SLOT_C),
		submmitBtn,
	}
	return components.FeedColumn(editor, editorID)
}

func (gs *GenState) PromptInput(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		if r.ParseForm() != nil {
			err := fmt.Errorf("bad form")
			InformError(err, &w)
			return
		}

		for slot, data := range r.Form {
			prompt := strings.Join(data, "")
			if _, ok := gs.prompts[slot]; ok {
				gs.prompts[slot] = prompt
			} else {
				err := fmt.Errorf("bad prompt")
				InformError(err, &w)
				return
			}
		}
	}
	w.WriteHeader(200)
}

// show all results and editor
func (gs *GenState) GenPage(w http.ResponseWriter, r *http.Request) {
	SetContentType(w, ContentType_Html)

	const colNum = 4
	var imgNum = len(gs.imageIds)
	var rowNum = int(math.Ceil(float64(imgNum) / colNum))
	// fmt.Println("+++ rows ", rowNum, " cols ", colNum)

	rows := make([]templ.Component, 0, rowNum)
	// row := make([]templ.Component, colNum)

	prevBtn := components.Pixelart()
	var lastImage templ.Component
	var images []templ.Component
	var drawImages int = 0
	for _, imgId := range gs.imageIds {
		if imgId == "deleted" {
			continue
		}

		drawImages += 1
		fmt.Printf("adding image: %d\n", drawImages)
		previewLink := JoinPath(PreviewOpen().Prefix, imgId)
		forPreview := UniqueModal(previewLink)
		forPreviewBtn := components.ModalButton(forPreview, prevBtn)
		forPreviewBtn1 := components.ModalButton(forPreview, prevBtn)
		lastImage = components.JustImg(imgUrl(imgId), imgDelUrl(imgId), forPreviewBtn, forPreviewBtn1)
		images = append(images, lastImage)

	}
	// delta := imgNum - showedImgNum
	// if delta > 0 {
	// 	rows = append(rows, components.FlexRow(row[0:delta]))
	// }
	var imgsLeft int = len(images)
	fmt.Printf("imgs %d\n", imgsLeft)
	fmt.Printf("rows %d\n", len(rows))

	var off int = 0
	for {
		if imgsLeft <= 4 {
			start := off
			end := off + imgsLeft
			rows = append(rows, components.FlexRow(images[start:end]))
			break
		}
		rows = append(rows, components.FlexRow(images[off:off+4]))
		imgsLeft -= 4
		off += 4
	}

	rows = append(rows, gs.PromptEditor("prompt_editor"))
	feed := components.FeedColumn(rows, "imgs")
	_ = feed
	wrap := components.FeedColumn([]templ.Component{
		feed,
		components.ModalIsland(UniqueModal("")),
	}, "modal_feed")
	wholePage := PageWithSidebar(wrap)
	wholePage.Render(context.Background(), w)
}

func (gs *GenState) PromptCommit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "!!! commit on get", 500)
	}

	for k, v := range gs.prompts {
		fmt.Println("+++ kv:", k, v)
	}

	// tu bym mógł zapisać prompty do bazy danych
	// może jako proto jako blob binarny w sql lite?
	id, err := imageGen(gs, gs.comfyData)
	if err != nil {
		log.Default().Println("ERR: ", err.Error())
		http.Error(w, "", 500)
		return
	}

	SetContentType(w, ContentType_Html)
	something := components.Block(888)
	feed := components.FeedColumn(
		[]templ.Component{
			components.JustImg(imgUrl(id), imgDelUrl(id), something, something),
			gs.PromptEditor("prompt_editor"),
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

	SetContentType(w, ContentType_Html)
	_, err := w.Write(imgData)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func InformError(err error, w *http.ResponseWriter) {
	if w != nil {
		http.Error(*w, err.Error(), 500)
	}
	log.Default().Println(err.Error())
}

func (ps *GenState) DeleteImage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if _, ok := ps.imageData[id]; !ok {
		InformError(fmt.Errorf("data for %s not present", id), &w)
		return
	}

	imgFile := JoinPath(DirImage(), Filename(id, "png"))
	if _, err := os.Stat(imgFile); err != nil {
		InformError(fmt.Errorf("file for %s dont exist | %v", id, err), &w)
		return
	}

	if err := os.RemoveAll(imgFile); err != nil {
		InformError(fmt.Errorf("%s file deletion failed", id), &w)
		return
	}

	delete(ps.imageData, id)
	if idx, ok := slices.BinarySearch(ps.imageIds, id); ok {
		ps.imageIds[idx] = "deleted"
		fmt.Printf("marked as deleted\n")
	}
	fmt.Println("+++ succesfully deleted: ", imgFile)
	w.WriteHeader(200)
}

func (ps *GenState) PreviewClose(w http.ResponseWriter, r *http.Request) {
	SetContentType(w, ContentType_Html)
	w.Write([]byte(`<div id="modal"></div>`))
}

func (ps *GenState) PreviewOpen(w http.ResponseWriter, r *http.Request) {
	var id string

	SetContentType(w, ContentType_Html)
	if id = r.PathValue("id"); id == "" {
		InformError(fmt.Errorf("failed to extract id"), &w)
	}

	fmt.Printf("+++ opening preview for %s\n", id)
	content := components.BigImg(imgUrl(id))
	modal := components.Modal("preview", content, PreviewClose().EntryPoint)
	modal.Render(r.Context(), w)
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

type LinkBind struct {
	Prefix     string
	EntryPoint string
}

func PreviewOpen() LinkBind {
	return LinkBind{
		Prefix:     "/preview/open",
		EntryPoint: "/preview/open/{id}",
	}
}

func PreviewClose() LinkBind {
	return LinkBind{
		Prefix:     "/preview/close",
		EntryPoint: "/preview/close/{id}",
	}
}

func (gs *GenState) LoadFns() HttpFuncMap {
	return HttpFuncMap{
		"/gen_page":                              gs.GenPage,
		"/prompt":                                gs.PromptInput,
		"/prompt/commit":                         gs.PromptCommit,
		"/prompt/img":                            gs.FetchImage,
		"/prompt/img/del/{id}":                   gs.DeleteImage,
		templ.SafeURL(PreviewOpen().EntryPoint):  gs.PreviewOpen,
		templ.SafeURL(PreviewClose().EntryPoint): gs.PreviewClose,
	}
}
