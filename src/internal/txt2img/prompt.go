package txt2img

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	mea_gen_d "mea_go/src/api/mea.gen.d"
	"mea_go/src/internal"
	utils "mea_go/src/internal"
	"mea_go/src/internal/translte"
	"net/http"
	"os"
	"strings"
	"sync"

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
	toTranslate []bool
	comfyData   *ComfyData
	otherState  *OtherState
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
	// fmt.Printf("|formated link - %s\n", formatedLink)
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

func PromptTranslateInitLB() LinkBind {
	return LinkBind{
		Prefix:     "/prompt/translate/init",
		EntryPoint: "/prompt/translate/init/{slot}",
		FmtStr:     "/prompt/translate/init/%s",
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
func ImgGenSink() HtmxId {
	return NamedId("image_sink")
}

func InvokeModal2(img2StartsWith string) internal.ActionLink {
	var open LinkBind = PreviewOpen()
	previewLink := open.FmtLink(img2StartsWith)
	modalHid := ModalHid()
	return internal.ActionLink{
		LinkToAction: previewLink,
		Target:       modalHid.TargName,
	}
}

var memory GenState

func init() {
	fmt.Println("+++ inital call")
	memory.init()
}

var emptyMetadata = ImgMetadata{prompts: []string{"", "", ""}}

// hmm maybe simple squil would be adequate, like SQLite or Postgress
func (gs *GenState) addImage(id string, gImg *image.RGBA, prompts SlotPromptS) error {
	var err error
	var pngBfr = bytes.Buffer{}

	err = png.Encode(&pngBfr, gImg)
	if err != nil {
		return fmt.Errorf("png encode failed | %w", err)
	}

	//saving image
	var dirImg = utils.DirImage()

	pngFile := utils.JoinPath(dirImg, utils.PngFilename(id))
	if err := data2File(pngFile, pngBfr); err != nil {
		return fmt.Errorf("!!! failed to save | %w", err)
	}

	yamlFile := utils.JoinPath(dirImg, utils.YamlFilename(id))
	err = utils.SaveAsYAML(yamlFile, prompts)
	if err != nil {
		return fmt.Errorf("!!! failed to save | %w", err)
	}
	metaD := ImgMetadata{prompts: prompts.Sequence}
	gs.otherState.addImg(id, pngBfr.Bytes(), metaD)
	return nil
}

func (gs *GenState) init() {
	gs.promptSlots = []mea_gen_d.Slot{
		mea_gen_d.Slot_a,
		mea_gen_d.Slot_b,
		mea_gen_d.Slot_c,
	}

	size := len(gs.promptSlots)
	gs.prompts = make(PromptMap, size)
	gs.toTranslate = make([]bool, size)
	for i := range size {
		gs.toTranslate[i] = true
	}
	for _, slot := range memory.promptSlots {
		gs.prompts[slot] = ""
	}

	var loadeImgsNum int = 0
	logger := log.Default()
	gs.otherState = loadOtherState(logger)
	logger.Println("INFO: ", fmt.Sprintf("loaded (%d) imgs on init", loadeImgsNum))
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

	genSink := ImgGenSink()
	submmitBtn := GenButton(genSink)

	editor := []templ.Component{
		padFromSlot(SLOT_A, mea_gen_d.Slot_a),
		padFromSlot(SLOT_B, mea_gen_d.Slot_b),
		padFromSlot(SLOT_C, mea_gen_d.Slot_c),
		submmitBtn,
	}
	return internal.FeedColumn(editor, hid.JustName)
}
func (gs *GenState) PromptTranslateInit(w http.ResponseWriter, r *http.Request) {
	lb := PromptTranslateLB()
	slot_name := r.PathValue("slot")

	slot, ok := SlotMapping[slot_name]
	if !ok {
		InformError(fmt.Errorf("bad slot 1"), w)
		return
	}

	htmlContent(w)
	if !gs.toTranslate[slot] {
		Tokens([]templ.Component{Token("Translation disabled")}).Render(r.Context(), w)
		return
	}
	link := lb.FmtLink(slot_name)
	err := SseReciver(link, "token").Render(r.Context(), w)
	if err != nil {
		InformError(err, w)
		return
	}
}

func compAsEvent(w io.Writer, evName string, comp templ.Component) error {
	var event bytes.Buffer
	fmt.Fprintf(&event, "event: %s\ndata:", evName)
	err := comp.Render(context.Background(), &event)
	if err != nil {
		return fmt.Errorf("render failed")
	}
	fmt.Fprintf(&event, "\n\n")

	_, err = w.Write(event.Bytes())
	if err != nil {
		return fmt.Errorf("failed while writing event")
	}
	return nil
}

func (gs *GenState) PromptTranslate(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		InformError(fmt.Errorf("sse not supported"), w)
		return
	}
	slot := r.PathValue("slot")

	slotKey, ok := SlotMapping[slot]
	if !ok {
		InformError(fmt.Errorf("bad slot 2"), w)
		return
	}

	prompt := gs.prompts[slotKey]

	// lets build prototype
	words := strings.Split(prompt, " ")
	spans := make([]templ.Component, 0, len(words))

	var td = 2000 / len(words)
	if td > 200 {
		td = 200
	}

	sseContent(w)
	internal.SetCacheControl(w, internal.CacheType_NoCache)
	w.Header().Set("Connection", "keep-alive")

	job := translte.TranlateJob{
		ToTranslate: prompt,
	}

	fullResponse := make([]string, 0, 64)
	tokenChan := make(chan string, 64)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for token := range tokenChan {
			fullResponse = append(fullResponse, token)
			spans = append(spans, Token(token))
			event := Tokens(spans)
			err := compAsEvent(w, "token", event)
			if err != nil {
				InformError(fmt.Errorf("falied at %d | %w", len(spans), err), w)
				return
			}
			flusher.Flush()
		}
	}()

	err := job.StreamedTranslateion(r.Context(), tokenChan)
	wg.Wait()
	if err != nil {
		InformError(fmt.Errorf("!!! translation failed | %w", err), w)
		return
	}

	fullText := strings.Join(fullResponse, "")
	gs.prompts[slotKey] = fullText
	fmt.Printf("full resonese was: %s", fullText)
	fmt.Fprintf(w, "event: done\ndata:\n\n")
	flusher.Flush()
}

func (gs *GenState) PromptInput(w http.ResponseWriter, r *http.Request) {
	var translate = PromptTranslateInitLB()

	if r.Method == http.MethodPost {
		if r.ParseForm() != nil {
			err := fmt.Errorf("bad form")
			InformError(err, w)
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
				out := internal.ProcedeNextVisible(transAction)
				// out = internal.Block(len(new))

				err := out.Render(r.Context(), w)
				if err != nil {
					log.Println(err.Error())
				}

				return
			} else {
				err := fmt.Errorf("bad prompt")
				InformError(err, w)
				return
			}
		}
	}
	w.WriteHeader(200)
}

func ImgComp(idImg string) templ.Component {
	var del LinkBind = ImageDelete()
	aLink := internal.ActionLink{
		LinkToAction: del.FmtLink(idImg),
		IDName:       fmt.Sprintf("deleter_%s", idImg),
		Target:       DeleteSinkHid().TargName,
	}
	aLink2 := InvokeModal2(idImg)
	forPreviewBtn := internal.ButtonAction(aLink2, internal.PixelartHold())

	delBtn := internal.ButtonAction(aLink, internal.Pixelart())
	return internal.JustImg(imgUrl(idImg), forPreviewBtn, delBtn)
}

type TCmpt = templ.Component

func (gs *GenState) PromptCommit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "!!! commit on get", 500)
	}

	for k, v := range gs.prompts {
		fmt.Println("+++ kv:", k, v)
	}

	// tu bym mógł zapisać prompty do bazy danych
	// może jako proto jako blob binarny w sql lite?
	newImgId, err := ImageGen(gs, gs.comfyData)
	if err != nil {
		log.Default().Println("ERR: ", err.Error())
		http.Error(w, "", 500)
		return
	}

	// ta := internal.TargetAction{
	// 	Target:       ImgGenSink().TargName,
	// 	LinkToAction: promptInput.FmtLink(id),
	// }

	spot := gs.otherState.placeInNewSpot(newImgId)
	_ = spot

	internal.SetContentType(w, internal.ContentType_Html)
	wrapped := IdWrap(ImgGenSink(), ImgComp(newImgId))
	wrapped.Render(context.Background(), w)

}

func (ps *GenState) FetchImage(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	id, ok := vals["id"]
	if !ok {
		http.Error(w, "bad request", 500)
		return
	}

	ops := ps.otherState
	imgData, ok := ops.imageData[id[0]]
	if !ok {
		http.Error(w, "bad request", 500)
		return
	}

	internal.SetContentType(w, internal.ContentType_Html)
	_, err := w.Write(imgData.bytes)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (ps *GenState) DeleteImage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	ops := ps.otherState
	if _, ok := ops.imageData[id]; !ok {
		InformError(fmt.Errorf("data for %s not present", id), w)
		return
	}

	exts := []string{"png", "yaml"}
	for _, ext := range exts {
		imgFile := internal.JoinPath(internal.DirImage(), internal.Filename(id, ext))
		if _, err := os.Stat(imgFile); err != nil {
			InformError(fmt.Errorf("file for %s dont exist | %v", id, err), w)
			return
		}

		if err := os.RemoveAll(imgFile); err != nil {
			InformError(fmt.Errorf("%s file deletion failed", id), w)
			return
		}
	}

	ops.deleteImg(id)
	fmt.Printf("+++ succesfully deleted: %s\n", id)

	htmlContent(w)

	// internal.ProcedeNext()
	IdWrap(DeleteSinkHid(), internal.Entry("del")).Render(r.Context(), w)
}

// show all results and editor
func (gs *GenState) GenPage(w http.ResponseWriter, r *http.Request) {
	internal.SetContentType(w, internal.ContentType_Html)

	const colNum = 4
	ogs := gs.otherState

	col := make([]templ.Component, 0, 4)
	for y := range 4 {
		row := make([]templ.Component, 0, 4)
		for x := range 4 {
			name := spotName(x, y)

			spotId := ogs.spotHolder[name]
			// fmt.Printf("- spot %s has id %s\n", name, spotId)
			var choice templ.Component
			if spotId != emptySpot {
				choice = ImgComp(spotId)
			} else {
				choice = internal.EmptyImgSlot()
			}

			row = append(row, IdWrap(NamedId(name), choice))
		}
		col = append(col, internal.FlexRow(row))
	}

	// it will become image matrix
	imgs := internal.FeedColumn(col, "imgs")
	mainContent := internal.FeedColumn([]templ.Component{
		gs.PromptEditor(EditorHid()),
		imgs,
		internal.ModalLayer(ModalHid()),
	}, "modal_feed")

	var opts = internal.PageOpts{
		PageContent: mainContent,
		Sinks: []HtmxId{
			DeleteSinkHid(),
			TranslateSinkHid(),
			ImgGenSink(),
		},
	}
	wholePage := internal.PageWithSidebar(opts)
	wholePage.Render(context.Background(), w)
}

func InformError(err error, w http.ResponseWriter) {
	if w != nil {
		http.Error(w, err.Error(), 500)
	}
	log.Default().Println(err.Error())
}

func htmlContent(w http.ResponseWriter) {
	internal.SetContentType(w, internal.ContentType_Html)
}
func sseContent(w http.ResponseWriter) {
	internal.SetContentType(w, internal.ContentType_EventStream)
}

func (ps *GenState) PreviewClose(w http.ResponseWriter, r *http.Request) {
	htmlContent(w)
	w.Write([]byte(`<div id="modal"></div>`))
}

func (ps *GenState) PreviewOpen(w http.ResponseWriter, r *http.Request) {
	var id string

	htmlContent(w)
	if id = r.PathValue("id"); id == "" {
		InformError(fmt.Errorf("failed to extract id"), w)
	}

	fmt.Printf("+++ opening preview for %s\n", id)
	if v, ok := ps.otherState.imageData[id]; ok {
		text := strings.Join(v.meta.prompts, " | ")
		fmt.Printf("+++ some data: %s\n", text)
	}

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
		templ.SafeURL(PromptInputLB().EntryPoint):         {Fn: gs.PromptInput, Show: true},
		templ.SafeURL(PromptTranslateInitLB().EntryPoint): {Fn: gs.PromptTranslateInit, Show: true},
		templ.SafeURL(PromptTranslateLB().EntryPoint):     {Fn: gs.PromptTranslate, Show: true},
		"/prompt/commit":                         {Fn: gs.PromptCommit, Show: true},
		"/prompt/img":                            {Fn: gs.FetchImage, Show: false},
		templ.SafeURL(ImageDelete().EntryPoint):  {Fn: gs.DeleteImage, Show: false},
		templ.SafeURL(PreviewOpen().EntryPoint):  {Fn: gs.PreviewOpen, Show: false},
		templ.SafeURL(PreviewClose().EntryPoint): {Fn: gs.PreviewClose, Show: false},
	}
}
