package internal

import (
	"context"
	"fmt"
	"mea_go/components"
	"net/http"
	"os"
	"strings"

	"github.com/a-h/templ"
)

var HeaderContentType = "Content-Type"

const (
	ContentTypeHtml        = "text/html"
	ContentTypeEventStream = "text/event-stream"
	ContentTypePng         = "image/png"
)

const feedId = "feedID"

type PromptMap = map[string]string
type HttpFuncMap = map[templ.SafeURL]HttpFunc

type GenState struct {
	prompts     PromptMap
	promptSlots []string
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
	for _, slot := range memory.promptSlots {
		gs.prompts[slot] = "placeholder"
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

func (ps *GenState) PromptFn(w http.ResponseWriter, r *http.Request) {

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
	editor := PromptEditor()
	editor.Render(context.Background(), w)
}

func (ps *GenState) PromptCommit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "!!! commit on get", 500)
	}

	for k, v := range ps.prompts {
		fmt.Println("+++", k, v)
	}

	w.Header().Set(HeaderContentType, ContentTypeHtml)
	feed := components.FeedColumn(
		[]templ.Component{
			components.JustImg(),
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

func (ps *GenState) PromptImage(w http.ResponseWriter, r *http.Request) {
	if len(img) == 0 {
		imgData, err := loadImage()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Println("odczytano zdjęcie")
		img = imgData
	}
	w.Header().Set(HeaderContentType, ContentTypePng)
	size, err := w.Write(img)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Println("+++ wysłano zdjęcie ", size)

}

func PromptSteteBlobalAcces() *GenState {
	return &memory
}

func (gs *GenState) LoadFns() HttpFuncMap {
	return HttpFuncMap{
		"/prompt":        gs.PromptFn,
		"/prompt/commit": gs.PromptCommit,
		"/prompt/img":    gs.PromptImage,
	}
}
