package components

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

var local PromptState
var HeaderContentType = "Content-Type"

const (
	ContentTypeHtml        = "text/html"
	ContentTypeEventStream = "text/event-stream"
)

type PromptState struct {
	prompts [3]string
}

func (ps *PromptState) setPrompt(id int, text string) {

}

func (ps *PromptState) PromptFn(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		if r.ParseForm() != nil {
			http.Error(w, "!!! not specified", 500)
		}

		for k, v := range r.Form {
			joined := strings.Join(v, "")
			fmt.Printf("+++ key: %s, value %s", k, joined)
			ps.setPrompt(0, joined)
		}
	}

	w.Header().Set(HeaderContentType, ContentTypeHtml)
	elem := PromptPad("o")
	elem.Render(context.Background(), w)
}

func PromptSteteBlobalAcces() *PromptState {
	return &local
}
