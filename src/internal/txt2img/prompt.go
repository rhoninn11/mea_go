package txt2img

import (
	"context"
	"fmt"
	"log"
	"math"
	mea_gen_d "mea_go/src/api/mea.gen.d"
	"mea_go/src/internal"
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

type SlotMap map[string]mea_gen_d.Slot

var SlotMapping = map[string]mea_gen_d.Slot{
	SLOT_A: mea_gen_d.Slot_a,
	SLOT_B: mea_gen_d.Slot_b,
	SLOT_C: mea_gen_d.Slot_c,
}

type HttpFuncMap = map[templ.SafeURL]internal.HttpFuncPack

type PromptMap = map[mea_gen_d.Slot]string

type ImgMetadata struct {
	prompts []string `yaml:"prompts"`
}
type ImgData struct {
	bytes []byte
	meta  ImgMetadata
}
type ImgMap = map[string]ImgData

type PromptSlots = []string

type GenState struct {
	prompts     PromptMap
	promptSlots []mea_gen_d.Slot
	comfyData   *ComfyData
	imageData   ImgMap
	imageIds    []string
	system      *any
}

type FlowData struct {
	slots PromptSlots
	rowId string
	colId string
}

type LinkBind struct {
	Prefix     string
	EntryPoint string
	FmtStr     string
}

func (lb *LinkBind) FmtLink(hmm ...any) string {
	formatedLink := fmt.Sprintf(lb.FmtStr, hmm...)
	fmt.Printf("|formated link - %s\n", formatedLink)
	return formatedLink

}

func PreviewOpen() LinkBind {
	return LinkBind{
		Prefix:     "/preview/open",
		EntryPoint: "/preview/open/{id}",
		FmtStr:     "/preview/open/%s",
	}
}

func PreviewClose() LinkBind {
	return LinkBind{
		Prefix:     "/preview/close",
		EntryPoint: "/preview/close/{id}",
	}
}

func ImageDelete() LinkBind {
	return LinkBind{
		Prefix:     "/prompt/img/del",
		EntryPoint: "/prompt/img/del/{id}",
		FmtStr:     "/prompt/img/del/%s",
	}
}

func PromptInputLB() LinkBind {
	return LinkBind{
		Prefix:     "/prompt/input",
		EntryPoint: "/prompt/input/{slot}",
		FmtStr:     "/prompt/input/%s",
	}
}

func PromptTranslateLB() LinkBind {
	return LinkBind{
		Prefix:     "/prompt/translate",
		EntryPoint: "/prompt/translate/{slot}",
		FmtStr:     "/prompt/translate/%s",
	}
}

type HtmxId = internal.HtmxId
type ModalDesc = internal.ModalDesc

func NamedId(name string) HtmxId {
	return HtmxId{
		JustName: name,
		TargName: fmt.Sprintf("#%s", name),
	}
}

func ModalHid() HtmxId {
	return NamedId("modal")
}

func EditorHid() HtmxId {
	return NamedId("prompt_editor")
}

func DeleteSinkHid() HtmxId {
	return NamedId("delete_sink")
}
func TranslateSinkHid() HtmxId {
	return NamedId("translate_sink")
}

func InvokeModal(link string) ModalDesc {
	return ModalDesc{
		Hid:      ModalHid(),
		WithLink: link,
	}
}

var memory GenState

func init() {
	fmt.Println("+++ inital call")
	memory.init()
}

var emptyMetadata = ImgMetadata{prompts: []string{"", "", ""}}

// hmm maybe simple squil would be adequate, like SQLite or Postgress
func (gs *GenState) addImage(id string, data []byte) {
	gs.imageIds = append(gs.imageIds, id)
	gs.imageData[id] = ImgData{
		meta:  emptyMetadata,
		bytes: data,
	}
}

func (gs *GenState) init() {
	gs.promptSlots = []mea_gen_d.Slot{
		mea_gen_d.Slot_a,
		mea_gen_d.Slot_b,
		mea_gen_d.Slot_c,
	}

	size := len(gs.promptSlots)
	gs.prompts = make(PromptMap, size)
	gs.imageData = make(ImgMap, 128)
	gs.imageIds = make([]string, 0, 128)
	for _, slot := range memory.promptSlots {
		gs.prompts[slot] = ""
	}

	var loadeImgsNum int = 0
	func(gs *GenState) {
		imgDir := internal.DirImage()
		entries, err := os.ReadDir(imgDir)
		if err != nil {
			log.Fatalln("scaning imgs", err.Error())
		}
		panicker := internal.Panicker(4)
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".png") {
				continue
			}
			basename := strings.TrimSuffix(name, ".png")
			yamlFile := internal.Filename(basename, "yaml")
			_ = yamlFile

			imgPath := strings.Join([]string{imgDir, name}, "/")
			metadataPath := strings.Join([]string{imgDir, name}, "/")

			var xd = ImgMetadata{prompts: []string{"", "", ""}}
			if _, err := os.Stat(metadataPath); err != nil {

			}

			data, err := os.ReadFile(imgPath)
			if panicker.HasError(err) {
				continue
			}

			loadeImgsNum += 1
			gs.imageIds = append(gs.imageIds, basename)
			gs.imageData[basename] = ImgData{
				bytes: data,
				meta:  xd,
			}
		}
	}(gs)

	log.Default().Println("INFO: ", fmt.Sprintf("loaded (%d) imgs on init", loadeImgsNum))
}

func (gs *GenState) PromptEditor(hid HtmxId) templ.Component {
	var promptInput = PromptInputLB()

	// dziewiąta tablica gilgameszha
	var sink = TranslateSinkHid()
	padFromSlot := func(id string, slot mea_gen_d.Slot) templ.Component {
		currPrompt := gs.prompts[slot]
		ta := internal.TargetAction{
			Target:       sink.TargName,
			LinkToAction: promptInput.FmtLink(id),
		}
		return PromptPadV2(id, currPrompt, ta)
	}

	submmitBtn := GenButton(hid.TargName)

	editor := []templ.Component{
		padFromSlot(SLOT_A, mea_gen_d.Slot_a),
		padFromSlot(SLOT_B, mea_gen_d.Slot_b),
		padFromSlot(SLOT_C, mea_gen_d.Slot_c),
		submmitBtn,
	}
	return internal.FeedColumn(editor, hid.JustName)
}
func (gs *GenState) PromptTranslate(w http.ResponseWriter, r *http.Request) {
	slot := r.PathValue("slot")

	fmt.Printf("+++ we will be translating for slot %s\n", slot)

	slotKey, ok := SlotMapping[slot]
	if !ok {
		fmt.Printf("+++ slot neveer existed %s\n", slot)
	}
	prompt := gs.prompts[slotKey]
	fmt.Printf("+++ current prompt %s\n", prompt)
}

func (gs *GenState) PromptInput(w http.ResponseWriter, r *http.Request) {
	var translate = PromptTranslateLB()

	if r.Method == http.MethodPost {
		if r.ParseForm() != nil {
			err := fmt.Errorf("bad form")
			InformError(err, &w)
			return
		}

		for slotName, data := range r.Form {

			prompt := strings.Join(data, "")
			fmt.Printf("%s | %s\n", slotName, prompt)
			slot := SlotMapping[slotName]

			if _, ok := gs.prompts[slot]; ok {
				old := gs.prompts[slot]
				new := prompt
				gs.prompts[slot] = new
				_ = old

				translationHid := NamedId(fmt.Sprintf("%s_pl", slotName))
				transAction := internal.TargetAction{
					Target:       translationHid.TargName,
					LinkToAction: translate.FmtLink(slotName),
				}
				out := internal.ProcedeNext(transAction)
				// out = internal.Block(len(new))

				err := out.Render(r.Context(), w)
				if err != nil {
					log.Println(err.Error())
				}

				return
			} else {
				err := fmt.Errorf("bad prompt")
				InformError(err, &w)
				return
			}
		}
	}
	w.WriteHeader(200)
}

func ImgComp(idImg string) templ.Component {
	asciiOpen := internal.PixelartHold()
	asciiDel := internal.Pixelart()
	var open LinkBind = PreviewOpen()
	var del LinkBind = ImageDelete()

	previewLink := open.FmtLink(idImg)
	modalDesc := InvokeModal(previewLink)
	forPreviewBtn := internal.ModalButton(modalDesc, asciiOpen)

	aLink := internal.ActionLink{
		LinkToAction: del.FmtLink(idImg),
		IDName:       fmt.Sprintf("deleter_%s", idImg),
		Target:       DeleteSinkHid().TargName,
	}
	delBtn := internal.ButtonAction(aLink, asciiDel)
	return internal.JustImg(imgUrl(idImg), imgDelUrl(idImg), forPreviewBtn, delBtn)
}

type TCmpt = templ.Component

// show all results and editor
func (gs *GenState) GenPage(w http.ResponseWriter, r *http.Request) {
	internal.SetContentType(w, internal.ContentType_Html)

	const colNum = 4
	var imgNum = len(gs.imageIds)
	var rowNum = int(math.Ceil(float64(imgNum) / colNum))
	// fmt.Println("+++ rows ", rowNum, " cols ", colNum)

	rows := make([]templ.Component, 0, rowNum)
	// row := make([]templ.Component, colNum)

	var lastImage templ.Component
	var images []templ.Component
	var drawImages int = 0

	for _, idImg := range gs.imageIds {
		fmt.Printf("|id image - %s\n", idImg)
		if idImg == "deleted" {
			continue
		}

		drawImages += 1
		lastImage = ImgComp(idImg)
		images = append(images, lastImage)

	}

	var imgsLeft int = len(images)
	fmt.Printf("imgs %d\n", imgsLeft)
	fmt.Printf("rows %d\n", len(rows))

	var off int = 0
	for {
		if imgsLeft <= 4 {
			start := off
			end := off + imgsLeft
			rows = append(rows, internal.FlexRow(images[start:end]))
			break
		}
		rows = append(rows, internal.FlexRow(images[off:off+4]))
		imgsLeft -= 4
		off += 4
	}
	// it will become image matrix
	slices.Reverse(rows)
	imgs := internal.FeedColumn(rows, "imgs")
	mainContent := internal.FeedColumn([]templ.Component{
		internal.JustHid(DeleteSinkHid()),
		gs.PromptEditor(EditorHid()),
		imgs,
		internal.ModalLayer(ModalHid()),
	}, "modal_feed")

	var opts = internal.PageOpts{
		PageContent: mainContent,
		Sinks: []HtmxId{
			DeleteSinkHid(),
			TranslateSinkHid(),
		},
	}
	wholePage := internal.PageWithSidebar(opts)
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
	id, err := ImageGen(gs, gs.comfyData)
	if err != nil {
		log.Default().Println("ERR: ", err.Error())
		http.Error(w, "", 500)
		return
	}

	internal.SetContentType(w, internal.ContentType_Html)
	something := internal.Block(888)

	promptId := NamedId("prompt_editor")
	feed := internal.FeedColumn(
		[]templ.Component{
			internal.JustImg(imgUrl(id), imgDelUrl(id), something, something),
			gs.PromptEditor(promptId),
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

	internal.SetContentType(w, internal.ContentType_Html)
	_, err := w.Write(imgData.bytes)
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

	imgFile := internal.JoinPath(internal.DirImage(), internal.Filename(id, "png"))
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
func htmlBrb(w http.ResponseWriter) {
	internal.SetContentType(w, internal.ContentType_Html)
}

func (ps *GenState) PreviewClose(w http.ResponseWriter, r *http.Request) {
	htmlBrb(w)
	w.Write([]byte(`<div id="modal"></div>`))
}

func (ps *GenState) PreviewOpen(w http.ResponseWriter, r *http.Request) {
	var id string

	htmlBrb(w)
	if id = r.PathValue("id"); id == "" {
		InformError(fmt.Errorf("failed to extract id"), &w)
	}

	fmt.Printf("+++ opening preview for %s\n", id)
	content := internal.BigImg(imgUrl(id))
	modal := internal.Modal("preview", content, PreviewClose().EntryPoint)
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

func (gs *GenState) LoadFns() HttpFuncMap {
	return HttpFuncMap{
		"/gen_page": {Fn: gs.GenPage, Show: true},
		templ.SafeURL(PromptInputLB().EntryPoint):     {Fn: gs.PromptInput, Show: false},
		templ.SafeURL(PromptTranslateLB().EntryPoint): {Fn: gs.PromptTranslate, Show: false},
		"/prompt/commit":                         {Fn: gs.PromptCommit, Show: false},
		"/prompt/img":                            {Fn: gs.FetchImage, Show: false},
		templ.SafeURL(ImageDelete().EntryPoint):  {Fn: gs.DeleteImage, Show: false},
		templ.SafeURL(PreviewOpen().EntryPoint):  {Fn: gs.PreviewOpen, Show: false},
		templ.SafeURL(PreviewClose().EntryPoint): {Fn: gs.PreviewClose, Show: false},
	}
}
