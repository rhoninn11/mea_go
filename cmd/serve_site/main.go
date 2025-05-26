package main

import (
	"context"
	"fmt"
	"log"
	"mea_go/components"
	"mea_go/internal"
	"net/http"
	"time"

	"github.com/a-h/templ"
)

type Direction struct {
	LatentVector string
	Power        float32
}

var mojUkladOdniesienia = map[string]Direction{
	"x": Direction{LatentVector: "Jeden z moich latent aspektów", Power: 0.3},
	"y": Direction{LatentVector: "Askpekt dominujący tego czego szukam", Power: 0.5},
	"z": Direction{LatentVector: "To takie moje oczko w głowie", Power: 0.2},
}

type State struct {
	keys      []string
	lastUsage int
	history   []string
}

var globState State

func init() {
	globState = State{
		keys:      []string{"x", "y", "z"},
		lastUsage: 0,
		history:   make([]string, 0, 16),
	}
}

var htmlType = "text/html"
var textStreamType = "text/event-stream"

func newGlobal(main templ.Component) templ.Component {
	side := components.SideLinks(endpotins)
	twoTabs := components.TwoTabs(side, main)
	return components.Global("Tua editro", twoTabs)
}

func (s *State) AxisFn(w ResponseWriter, r *Request) {
	dir := mojUkladOdniesienia[s.keys[s.lastUsage]]
	text := dir.LatentVector

	w.Header().Set("Content-Type", htmlType)
	historyNote := fmt.Sprintf("%s |może data|", text)
	entry := components.Entry(historyNote)
	glob := newGlobal(entry)
	glob.Render(r.Context(), w)

	s.lastUsage += 1
	s.history = append(s.history, historyNote)
	if s.lastUsage == len(s.keys) {
		s.lastUsage = 0
	}
}

func (s *State) HistoryFn(w ResponseWriter, r *Request) {
	w.Header().Set("Content-Type", htmlType)
	history := components.HistoryWhole(s.history)
	page := newGlobal(history)
	page.Render(context.Background(), w)
}

func (s *State) LoadingPage(w ResponseWriter, r *Request) {

	w.Header().Set("Content-Type", htmlType)
	render := components.SectionWithLoading()
	render = components.Global("Loading Page", render)
	render.Render(context.Background(), w)
}

// /page/gen
func (s *State) GeneratePage(w ResponseWriter, r *Request) {
	w.Header().Set("Content-Type", htmlType)
	promptPad := components.PromptPad("0")
	fullPage := newGlobal(promptPad)
	fullPage.Render(context.Background(), w)
}

func (s *State) RecivePrompt(w ResponseWriter, r *Request) {
	fmt.Println("+++wywołano well")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Fatal("empty form")
	}

	form := r.Form
	fmt.Printf("+++ form len: %d\n", len(form))
	for k, v := range form {
		fmt.Println("+++", k, v)
	}

	w.Header().Set("Content-Type", htmlType)
	render := components.Block(0)
	render.Render(context.Background(), w)
}

func loading(w ResponseWriter, r *Request) {
	w.Header().Set("Content-Type", textStreamType) // but i hope html can be pushed fine😅
	for i := range 10 {
		render := components.Block(i)
		render.Render(context.Background(), w)
		w.(http.Flusher).Flush()
		time.Sleep(10 * time.Millisecond)
	}
}

func spf(format string, a ...any) string {
	var temp = fmt.Sprintf(format, a...)
	// fmt.Println("+++Debug: ", temp)
	return temp
}

var endpotins = make([]templ.SafeURL, 0, 16)

type ResponseWriter = http.ResponseWriter
type Request = http.Request
type HttpFuncSignature = func(ResponseWriter, *Request)

func httpHandleFunc(endpoint templ.SafeURL, fn HttpFuncSignature) {
	endpotins = append(endpotins, endpoint)
	http.HandleFunc(string(endpoint), fn)
}

func main() {
	// for _, key := range keys {
	// 	fmt.Println(mojUkladOdniesienia[key])
	// }
	const host = "localhost"
	const port = 8080
	var base = spf("%s:%d", host, port) // eg localhost:8080

	static := http.FileServer(http.Dir("./static/"))

	// TODO: for static files rebuild in development would be nice to set
	// cache-control header to no cache somehow in dev mode
	http.Handle("/static/", http.StripPrefix("/static/", static))

	deeper := internal.PromptSteteBlobalAcces()

	httpHandleFunc("/axis", globState.AxisFn)
	httpHandleFunc("/history", globState.HistoryFn)
	httpHandleFunc("/loading", loading)
	httpHandleFunc("/page/loading", globState.LoadingPage)
	httpHandleFunc("/page/gen", globState.GeneratePage)
	httpHandleFunc("/well", globState.RecivePrompt)
	httpHandleFunc("/prompt", deeper.PromptFn)

	var url = spf("http://%s/%s", base, "history")
	fmt.Printf("+++ niby wystartowałem api, api route: \n%s\n", url)
	http.ListenAndServe(base, nil)
}
